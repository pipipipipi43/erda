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

package autotest

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var SPACE_EXPORT = apis.ApiSpec{
	Path:         "/api/autotests/spaces/actions/export",
	BackendPath:  "/api/autotests/spaces/actions/export",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodGet,
	RequestType:  apistructs.AutoTestSpaceExportRequest{},
	ResponseType: apistructs.AutoTestSpaceExportResponse{},
	CheckLogin:   true,
	Doc:          "summary: 导出自动化测试空间",
}
