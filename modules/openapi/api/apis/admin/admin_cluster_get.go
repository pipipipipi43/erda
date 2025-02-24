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

package admin

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ADMIN_CLUSTER_GET = apis.ApiSpec{
	Path:         "/api/clusters/<clusterName>",
	BackendPath:  "/api/clusters/<clusterName>",
	Method:       "GET",
	Host:         "admin.marathon.l4lb.thisdcos.directory:9095",
	Scheme:       "http",
	CheckLogin:   true,
	ResponseType: apistructs.ClusterInfo{},
	Doc:          "summary: 集群列表",
}
