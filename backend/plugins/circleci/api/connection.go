/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/circleci/models"
	"github.com/apache/incubator-devlake/server/api/shared"
)

type CircleciTestConnResponse struct {
	shared.ApiBody
}

func testConnection(ctx context.Context, connection models.CircleciConn) (*CircleciTestConnResponse, errors.Error) {
	// validate
	if vld != nil {
		if err := vld.Struct(connection); err != nil {
			return nil, errors.Default.Wrap(err, "error validating target")
		}
	}
	// test connection
	apiClient, err := api.NewApiClientFromConnection(ctx, basicRes, &connection)
	if err != nil {
		return nil, err
	}

	res, err := apiClient.Get("/v2/me", nil, nil)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.HttpStatus(res.StatusCode).New(fmt.Sprintf("unexpected status code: %d", res.StatusCode))
	}

	body := CircleciTestConnResponse{}
	body.Success = true
	body.Message = "success"
	// output
	return &body, nil
}

// TestConnection test circleci connection
// @Summary test circleci connection
// @Description Test circleci Connection
// @Tags plugins/circleci
// @Param body body models.CircleciConnection true "json body"
// @Success 200  {object} CircleciTestConnResponse "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/circleci/test [POST]
func TestConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	// process input
	var connection models.CircleciConn
	err := api.Decode(input.Body, &connection, vld)
	if err != nil {
		return nil, err
	}
	// test connection
	result, err := testConnection(context.TODO(), connection)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// TestExistingConnection test circleci connection
// @Summary test circleci connection
// @Description Test circleci Connection
// @Tags plugins/circleci
// @Success 200  {object} CircleciTestConnResponse "Success"
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/circleci/{connectionId}/test [POST]
func TestExistingConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.CircleciConnection{}
	err := connectionHelper.First(connection, input.Params)
	if err != nil {
		return nil, errors.BadInput.Wrap(err, "find connection from db")
	}
	if err := api.DecodeMapStruct(input.Body, connection, false); err != nil {
		return nil, err
	}
	// test connection
	result, err := testConnection(context.TODO(), connection.CircleciConn)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// PostConnections @Summary create circleci connection
// @Description Create circleci connection
// @Tags plugins/circleci
// @Param body body models.CircleciConnection true "json body"
// @Success 200  {object} models.CircleciConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/circleci/connections [POST]
func PostConnections(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	// update from request and save to database
	connection := &models.CircleciConnection{}
	err := connectionHelper.Create(connection, input)
	if err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: connection.Sanitize(), Status: http.StatusOK}, nil
}

// PatchConnection @Summary patch circleci connection
// @Description Patch circleci connection
// @Tags plugins/circleci
// @Param body body models.CircleciConnection true "json body"
// @Success 200  {object} models.CircleciConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/circleci/connections/{connectionId} [PATCH]
func PatchConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.CircleciConnection{}
	if err := connectionHelper.First(&connection, input.Params); err != nil {
		return nil, err
	}
	if err := (&models.CircleciConnection{}).MergeFromRequest(connection, input.Body); err != nil {
		return nil, errors.Convert(err)
	}
	if err := connectionHelper.SaveWithCreateOrUpdate(connection); err != nil {
		return nil, err
	}
	return &plugin.ApiResourceOutput{Body: connection.Sanitize()}, nil
}

// DeleteConnection @Summary delete a circleci connection
// @Description Delete a circleci connection
// @Tags plugins/circleci
// @Success 200  {object} models.CircleciConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/circleci/connections/{connectionId} [DELETE]
func DeleteConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	conn := &models.CircleciConnection{}
	output, err := connectionHelper.Delete(conn, input)
	if err != nil {
		return output, err
	}
	output.Body = conn.Sanitize()
	return output, nil

}

// ListConnections @Summary get all circleci connections
// @Description Get all circleci connections
// @Tags plugins/circleci
// @Success 200  {object} []models.CircleciConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/circleci/connections [GET]
func ListConnections(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	var connections []models.CircleciConnection
	err := connectionHelper.List(&connections)
	if err != nil {
		return nil, err
	}
	for idx, c := range connections {
		connections[idx] = c.Sanitize()
	}
	return &plugin.ApiResourceOutput{Body: connections, Status: http.StatusOK}, nil
}

// GetConnection @Summary get circleci connection detail
// @Description Get circleci connection detail
// @Tags plugins/circleci
// @Success 200  {object} models.CircleciConnection
// @Failure 400  {string} errcode.Error "Bad Request"
// @Failure 500  {string} errcode.Error "Internal Error"
// @Router /plugins/circleci/connections/{connectionId} [GET]
func GetConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.CircleciConnection{}
	err := connectionHelper.First(connection, input.Params)
	return &plugin.ApiResourceOutput{Body: connection.Sanitize()}, err
}
