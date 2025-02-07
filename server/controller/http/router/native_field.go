/*
 * Copyright (c) 2024 Yunshan Networks
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package router

import (
	"github.com/deepflowio/deepflow/server/controller/common"
	httpcommon "github.com/deepflowio/deepflow/server/controller/http/common"
	"github.com/deepflowio/deepflow/server/controller/http/common/response"
	"github.com/deepflowio/deepflow/server/controller/http/service"
	"github.com/deepflowio/deepflow/server/controller/model"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type NativeField struct{}

func NewNativeField() *NativeField {
	return new(NativeField)
}

func (a *NativeField) RegisterTo(e *gin.Engine) {
	adminRoutes := e.Group("/v1/native-fields")
	adminRoutes.Use(AdminPermissionVerificationMiddleware())

	adminRoutes.GET("/", getNativeFields)
	adminRoutes.POST("/:lcuuid/", createNativeField)
	adminRoutes.DELETE("/:lcuuid/", deleteNativeField)
}

func getNativeFields(c *gin.Context) {
	orgID, _ := c.Get(common.HEADER_KEY_X_ORG_ID)
	data, err := service.GetNativeFields(orgID.(int))
	response.JSON(c, response.SetData(data), response.SetError(err))
}

func createNativeField(c *gin.Context) {
	var err error
	var nativeFieldCreate model.NativeFieldCreate

	// 参数校验
	err = c.ShouldBindBodyWith(&nativeFieldCreate, binding.JSON)
	if err != nil {
		response.JSON(c, response.SetStatus(httpcommon.INVALID_PARAMETERS), response.SetDescription(err.Error()))
		return
	}
	orgID, _ := c.Get(common.HEADER_KEY_X_ORG_ID)
	data, err := service.CreateNativeField(orgID.(int), nativeFieldCreate)
	response.JSON(c, response.SetData(data), response.SetError(err))
}

func deleteNativeField(c *gin.Context) {
	lcuuid := c.Param("lcuuid")
	orgID, _ := c.Get(common.HEADER_KEY_X_ORG_ID)
	data, err := service.DeleteNativeField(orgID.(int), lcuuid)
	response.JSON(c, response.SetData(data), response.SetError(err))
}
