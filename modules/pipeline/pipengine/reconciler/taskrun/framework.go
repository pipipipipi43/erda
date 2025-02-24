// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package taskrun

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

func (tr *TaskRun) Do(itr TaskOp) error {
	logrus.Infof("reconciler: pipelineID: %d, task %q begin %s", tr.P.ID, tr.Task.Name, itr.Op())

	o := &Elem{ErrCh: make(chan error), DoneCh: make(chan interface{}), ExitCh: make(chan struct{})}
	o.TimeoutCh, o.Cancel, o.Timeout = itr.TimeoutConfig()

	// define op handle func
	handleProcessingResult := func(data interface{}, err error) {
		// fetchLatestTask for task update after processing
		_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Minute)).
			Do(func() (bool, error) { return tr.fetchLatestTask() == nil, nil })

		if tr.Task.Status.IsEndStatus() {
			o.ExitCh <- struct{}{}
			return
		}

		if err != nil {
			o.ErrCh <- err
			return
		}
		o.DoneCh <- data
		return
	}

	go func() {
		var err error
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				err = errors.Errorf("taskOp: %s, panic: %v", itr.Op(), r)
				handleProcessingResult(nil, err)
			}
		}()

		// fetch latest pipeline status to judge whether to continue do task
		_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Minute)).
			Do(func() (bool, error) { return tr.fetchLatestPipelineStatus() == nil, nil })
		if tr.QueriedPipelineStatus.IsEndStatus() {
			rlog.TWarnf(tr.P.ID, tr.Task.ID,
				"query latest pipeline status is already end status (%s), so stop reconciler task, current op: %s",
				tr.QueriedPipelineStatus, string(itr.Op()))
			tr.PExitChCancel()
			return
		}

		// aop: before processing
		if itr.TuneTriggers().BeforeProcessing != "" {
			_ = aop.Handle(aop.NewContextForTask(*tr.Task, *tr.P, itr.TuneTriggers().BeforeProcessing))
		}

		// processing op
		data, err := itr.Processing()

		// op handle processing result
		handleProcessingResult(data, err)
	}()

	return tr.waitOp(itr, o)
}

func (tr *TaskRun) waitOp(itr TaskOp, o *Elem) (result error) {
	var (
		// errs 表示任务异常，需要重试
		errs []string
		// resultErrMsg 仅记录到 task.result.errors，不表示任务异常
		resultErrMsg []string
	)
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			errs = append(errs, fmt.Sprintf("%v", r))
		}
		resultErrMsg = append(resultErrMsg, errs...)
		if len(resultErrMsg) > 0 {
			tr.Task.Result.Errors = append(tr.Task.Result.Errors, apistructs.ErrorResponse{Msg: strutil.Join(resultErrMsg, "\n", true)})
		}

		// loop
		if err := tr.handleTaskLoop(); err != nil {
			// append err loop
			errs = append(errs, fmt.Sprintf("%v", err))
		}

		// if we invoke `tr.fetchLatestTask` method here before `update`,
		// we will lost changes made by `WhenXXX` methods.
		tr.Update()

		if len(errs) > 0 {
			result = errors.Errorf("failed to %s task, err: %s", itr.Op(), strutil.Join(errs, "\n", true))
		}
	}()

	// timeout cancel might be nil
	if o.Cancel != nil {
		defer o.Cancel()
	}

	select {
	case <-tr.Ctx.Done():
		// 被外部取消
		tr.PExit = true
		rlog.TWarnf(tr.P.ID, tr.Task.ID, "received stop reconcile signal, canceled, reason: %s", tr.Ctx.Err())
		return

	case data := <-o.DoneCh:
		tr.LogStep(itr.Op(), "begin do WhenDone")
		defer tr.LogStep(itr.Op(), "end do WhenDone")
		if err := itr.WhenDone(data); err != nil {
			errs = append(errs, err.Error())
		}
		// aop
		_ = aop.Handle(aop.NewContextForTask(*tr.Task, *tr.P, itr.TuneTriggers().AfterProcessing))

	case err := <-o.ErrCh:
		logrus.Errorf("reconciler: pipelineID: %d, task %q %s received error (%v)", tr.P.ID, tr.Task.Name, itr.Op(), err)
		errs = append(errs, err.Error())
		tr.LogStep(itr.Op(), "begin do WhenLogicError")
		defer tr.LogStep(itr.Op(), "end do WhenLogicError")
		if err := itr.WhenLogicError(err); err != nil {
			errs = append(errs, err.Error())
		}

	case <-o.TimeoutCh:
		// 超时需要手动更新 task
		_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Minute)).
			Do(func() (bool, error) { return tr.fetchLatestTask() == nil, nil })

		tr.LogStep(itr.Op(), "begin do WhenTimeout")
		defer tr.LogStep(itr.Op(), "end do WhenTimeout")
		if err := itr.WhenTimeout(); err != nil {
			errs = append(errs, err.Error())
		}

		if itr.TaskRun().FakeTimeout {
			return
		}

		logrus.Errorf("reconciler: pipelineID: %d, task %q %s received timeout (%s)", tr.P.ID, tr.Task.Name, itr.Op(), o.Timeout)
		resultErrMsg = append(resultErrMsg, fmt.Sprintf("timeout (%s) (platform: %s)", o.Timeout, conf.TaskDefaultTimeout()))

	case <-o.ExitCh:
		// 说明 task 在处理过程中被外部流程（例如 取消流水线）已经置为终态，直接结束即可
		tr.LogStep(itr.Op(), "waitOp received ExitCh")
		return

	}

	return
}

// reconciler: pipelineID: 1, taskID: 1, taskName: repo, taskOp: start, step: begin do WhenDone
func (tr *TaskRun) LogStep(taskOp Op, step string) {
	logrus.Debugf("reconciler: pipelineID: %d, taskID: %d, taskName: %s, taskOp: %s, step: %s",
		tr.P.ID, tr.Task.ID, tr.Task.Name, string(taskOp), step)
}
