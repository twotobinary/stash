package manager

import (
	"os"
	"sync"

	"github.com/stashapp/stash/pkg/ffmpeg"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/manager/config"
	"github.com/stashapp/stash/pkg/models"
)

type GenerateTranscodeTask struct {
	Scene models.Scene
}

func (t *GenerateTranscodeTask) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	videoCodec := t.Scene.VideoCodec.String
	if ffmpeg.IsValidCodec(videoCodec) {
		return
	}

	hasTranscode, _ := HasTranscode(&t.Scene)
	if hasTranscode {
		return
	}

	logger.Infof("[transcode] <%s> scene has codec %s", t.Scene.Checksum, t.Scene.VideoCodec.String)

	videoFile, err := ffmpeg.NewVideoFile(instance.FFProbePath, t.Scene.Path)
	if err != nil {
		logger.Errorf("[transcode] error reading video file: %s", err.Error())
		return
	}

	outputPath := instance.Paths.Generated.GetTmpPath(t.Scene.Checksum + ".mp4")
	transcodeSize := config.GetMaxTranscodeSize()
	options := ffmpeg.TranscodeOptions{
		OutputPath:       outputPath,
		MaxTranscodeSize: transcodeSize,
	}
	encoder := ffmpeg.NewEncoder(instance.FFMPEGPath)
	encoder.Transcode(*videoFile, options)
	if err := os.Rename(outputPath, instance.Paths.Scene.GetTranscodePath(t.Scene.Checksum)); err != nil {
		logger.Errorf("[transcode] error generating transcode: %s", err.Error())
		return
	}
	logger.Debugf("[transcode] <%s> created transcode: %s", t.Scene.Checksum, outputPath)
	return
}

func (t *GenerateTranscodeTask) isTranscodeNeeded() bool {

	videoCodec := t.Scene.VideoCodec.String
	hasTranscode, _ := HasTranscode(&t.Scene)

	if ffmpeg.IsValidCodec(videoCodec) {
		return false
	}

	if hasTranscode {
		return false
	}
	return true
}
