package vulkan

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestVulkan_TranscodeToAvcCmd_WithDevice(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.VulkanAvc, 1500, encode.DefaultQuality, encode.PresetFast, "0", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-init_hw_device vulkan=vk:0"))
	assert.True(t, strings.Contains(s, "-filter_hw_device vk"))
	assert.True(t, strings.Contains(s, "format=nv12"))
	assert.True(t, strings.Contains(s, "hwupload"))
	assert.True(t, strings.Contains(s, "-c:v h264_vulkan"))
	assert.True(t, strings.Contains(s, "-qp 25"))
	assert.True(t, strings.Contains(s, "-f mp4"))
}

func TestVulkan_TranscodeToAvcCmd_NoDevice(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.VulkanAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-init_hw_device vulkan=vk "))
	assert.False(t, strings.Contains(s, "vulkan=vk:"))
	assert.True(t, strings.Contains(s, "-filter_hw_device vk"))
	assert.True(t, strings.Contains(s, "-c:v h264_vulkan"))
}
