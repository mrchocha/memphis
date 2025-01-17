// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"fmt"

	"github.com/memphisdev/memphis/models"

	"github.com/gin-gonic/gin"
)

type FunctionsHandler struct{}

func (fh FunctionsHandler) GetAllFunctions(c *gin.Context) {
	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllFunctions at getUserDetailsFromMiddleware: %v", err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	functionsResult, err := fh.GetFunctions(user.TenantName)
	if err != nil {
		serv.Errorf("[tenant: %v][user: %v]GetAllFunctions at GetFunctions: %v", user.TenantName, user.Username, err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	c.JSON(200, gin.H{"scm_integrated": functionsResult.ScmIntegrated, "functions": functionsResult.Functions})
}

func (fh FunctionsHandler) GetFunctions(tenantName string) (models.FunctionsRes, error) {
	contentDetailsOfSelectedRepos, scmIntegrated := GetContentOfSelectedRepos(tenantName)
	functions, err := GetFunctionsDetails(contentDetailsOfSelectedRepos)
	if err != nil {
		return models.FunctionsRes{}, err
	}
	allFunctions := models.FunctionsRes{
		Functions:     functions,
		ScmIntegrated: scmIntegrated,
	}

	return allFunctions, nil
}

func validateYamlContent(yamlMap map[string]interface{}) error {
	requiredFields := []string{"function_name", "language"}
	missingFields := make([]string, 0)
	for _, field := range requiredFields {
		if _, exists := yamlMap[field]; !exists {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("Missing fields: %v\n", missingFields)
	}
	return nil
}

func GetFunctionsDetails(functionsDetails []functionDetails) ([]models.FunctionsResult, error) {
	functions := []models.FunctionsResult{}
	for _, functionDetails := range functionsDetails {
		fucntionContentMap := functionDetails.ContentMap
		commit := functionDetails.Commit
		fileContent := functionDetails.Content
		repo := functionDetails.RepoName
		branch := functionDetails.Branch
		owner := functionDetails.Owner
		tagsInterfaceSlice, ok := fucntionContentMap["tags"].([]interface{})
		tagsStrings := []string{}
		if ok {
			tagsStrings = make([]string, len(fucntionContentMap["tags"].([]interface{})))
			for i, tag := range tagsInterfaceSlice {
				tagMap := tag.(map[interface{}]interface{})
				for _, v := range tagMap {
					if str, ok := v.(string); ok {
						tagsStrings[i] = str
					}
				}
			}
		}

		description, ok := fucntionContentMap["description"].(string)
		if !ok {
			description = ""
		}

		functionDetails := models.FunctionsResult{
			FunctionName: fucntionContentMap["function_name"].(string),
			Description:  description,
			Tags:         tagsStrings,
			Language:     fucntionContentMap["language"].(string),
			LastCommit:   *commit.Commit.Committer.Date,
			Link:         *fileContent.HTMLURL,
			Repository:   repo,
			Branch:       branch,
			Owner:        owner,
		}

		functions = append(functions, functionDetails)
	}
	return functions, nil
}
