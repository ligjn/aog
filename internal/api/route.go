// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/ligjn/aog/config"
	"github.com/ligjn/aog/internal/datastore"
	"github.com/ligjn/aog/internal/provider"
	"github.com/ligjn/aog/internal/types"
	"github.com/ligjn/aog/internal/utils"
	"github.com/ligjn/aog/version"
)

func InjectRouter(e *AOGCoreServer) {
	e.Router.Handle(http.MethodGet, "/", rootHandler)
	e.Router.Handle(http.MethodGet, "/health", healthHeader)
	e.Router.Handle(http.MethodGet, "/engine/health", engineHealthHandler)
	e.Router.Handle(http.MethodGet, "/version", getVersion)
	e.Router.Handle(http.MethodGet, "/engine/version", getEngineVersion)
	e.Router.Handle(http.MethodGet, "/update/status", updateAvailableHandler)
	e.Router.Handle(http.MethodPost, "/update", updateHandler)

	r := e.Router.Group("/aog/" + version.AOGVersion)

	// service import / export
	r.Handle(http.MethodPost, "/service/export", e.ExportService)
	r.Handle(http.MethodPost, "/service/import", e.ImportService)

	// Inject the router into the server
	r.Handle(http.MethodPost, "/service/install", e.CreateAIGCService)
	r.Handle(http.MethodPut, "/service", e.UpdateAIGCService)
	r.Handle(http.MethodGet, "/service", e.GetAIGCServices)

	r.Handle(http.MethodGet, "/service_provider", e.GetServiceProviders)
	r.Handle(http.MethodPost, "/service_provider", e.CreateServiceProvider)
	r.Handle(http.MethodPut, "/service_provider", e.UpdateServiceProvider)
	r.Handle(http.MethodDelete, "/service_provider", e.DeleteServiceProvider)

	r.Handle(http.MethodGet, "/model", e.GetModels)
	r.Handle(http.MethodPost, "/model", e.CreateModel)
	r.Handle(http.MethodDelete, "/model", e.DeleteModel)
	r.Handle(http.MethodPost, "/model/stream", e.CreateModelStream)
	r.Handle(http.MethodPost, "/model/stream/cancel", e.CancelModelStream)
	r.Handle(http.MethodGet, "/model/recommend", e.GetRecommendModels)
	r.Handle(http.MethodGet, "/model/support", e.GetModelList)

	slog.Info("Gateway started", "host", config.GlobalAOGEnvironment.ApiHost)
}

func rootHandler(c *gin.Context) {
	c.String(http.StatusOK, "AIPC Open Gateway")
}

func healthHeader(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "UP"})
}

func engineHealthHandler(c *gin.Context) {
	var data = make(map[string]string)
	for _, modelEngineName := range types.SupportModelEngine {
		engine := provider.GetModelEngine(modelEngineName)
		err := engine.HealthCheck()
		if err != nil {
			data[modelEngineName] = "DOWN"
			continue
		}
		data[modelEngineName] = "UP"
	}
	c.JSON(http.StatusOK, data)
}

func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"version": version.AOGVersion})
}

func getEngineVersion(c *gin.Context) {
	ctx := c.Request.Context()
	var data = make(map[string]string)
	for _, modelEngineName := range types.SupportModelEngine {
		engine := provider.GetModelEngine(modelEngineName)
		var respData types.EngineVersionResponse
		resp, err := engine.GetVersion(ctx, &respData)
		if err != nil {
			data[modelEngineName] = "get version failed"
			continue
		}
		data[modelEngineName] = resp.Version
	}

	c.JSON(http.StatusOK, data)
}

func updateAvailableHandler(c *gin.Context) {
	ctx := c.Request.Context()
	status, updateResp := version.IsNewVersionAvailable(ctx)
	if status {
		c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Ollama version %s is ready to install", updateResp.UpdateVersion)})
	} else {
		c.JSON(http.StatusOK, map[string]string{"message": ""})
	}

}

func updateHandler(c *gin.Context) {
	// check server
	status := utils.IsServerRunning()
	if status {
		// stop server
		pidFilePath := filepath.Join(config.GlobalAOGEnvironment.RootDir, "aog.pid")
		err := utils.StopAOGServer(pidFilePath)
		if err != nil {
			c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
		}
	}
	// rm old version file
	aogFileName := "aog.exe"
	if runtime.GOOS != "windows" {
		aogFileName = "aog"
	}
	aogFilePath := filepath.Join(config.GlobalAOGEnvironment.RootDir, aogFileName)
	err := os.Remove(aogFilePath)
	if err != nil {
		slog.Error("[Update] Failed to remove aog file %s: %v\n", aogFilePath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	// install new version
	downloadPath := filepath.Join(config.GlobalAOGEnvironment.RootDir, "download", aogFileName)
	err = os.Rename(downloadPath, aogFilePath)
	if err != nil {
		slog.Error("[Update] Failed to rename aog file %s: %v\n", downloadPath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	// start server
	logPath := config.GlobalAOGEnvironment.ConsoleLog
	rootDir := config.GlobalAOGEnvironment.RootDir
	err = utils.StartAOGServer(logPath, rootDir)
	if err != nil {
		slog.Error("[Update] Failed to start aog log %s: %v\n", logPath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	ds := datastore.GetDefaultDatastore()
	ctx := c.Request.Context()
	vr := &types.VersionUpdateRecord{}
	sortOption := []datastore.SortOption{
		{Key: "created_at", Order: -1},
	}
	versionRecoreds, err := ds.List(ctx, vr, &datastore.ListOptions{SortBy: sortOption})
	if err != nil {
		slog.Error("[Update] Failed to list versions: %v\n", err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	versionRecord := versionRecoreds[0].(*types.VersionUpdateRecord)
	if versionRecord.Status == types.VersionRecordStatusInstalled {
		versionRecord.Status = types.VersionRecordStatusUpdated
	}
	err = ds.Put(ctx, versionRecord)
	if err != nil {
		slog.Error("[Update] Failed to update versions: %v\n", err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	c.JSON(http.StatusOK, map[string]string{"message": ""})
}
