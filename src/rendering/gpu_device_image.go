/******************************************************************************/
/* gpu_device_image.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering/gpu_types"
	"kaijuengine.com/rendering/vulkan_const"
)

func (g *GPUDevice) CreateImage(id *TextureId, properties gpu_types.MemoryPropertyFlags, req GPUImageCreateRequest) error {
	defer tracing.NewRegion("GPUDevice.CreateImage").End()
	return g.createImageImpl(id, properties, req)
}

func (g *GPUDevice) CreateBasicTextureSampler(mipLevels uint32) (gpu_types.Sampler, error) {
	return g.CreateTextureSampler(mipLevels, vulkan_const.SamplerAddressModeRepeat, gpu_types.FilterLinear)
}

func (g *GPUDevice) CreateTextureSampler(mipLevels uint32, samplerMode vulkan_const.SamplerAddressMode, filter gpu_types.Filter) (gpu_types.Sampler, error) {
	defer tracing.NewRegion("GPULogicalDevice.CreateTextureSampler").End()
	return g.createTextureSamplerImpl(mipLevels, samplerMode, filter)
}

func (g *GPUDevice) TransitionImageLayout(vt *TextureId, newLayout gpu_types.ImageLayout, aspectMask gpu_types.ImageAspectFlags, newAccess gpu_types.AccessFlags, cmd *CommandRecorder) {
	defer tracing.NewRegion("GPUDevice.TransitionImageLayout").End()
	g.transitionImageLayoutImpl(vt, newLayout, aspectMask, newAccess, cmd)
}

func (g *GPUDevice) CopyBufferToImage(buffer gpu_types.Buffer, image gpu_types.Image, width, height uint32, layerCount int) {
	defer tracing.NewRegion("GPUDevice.CopyBufferToImage").End()
	g.copyBufferToImageImpl(buffer, image, width, height, layerCount)
}

func (g *GPUDevice) WriteBufferToImageRegion(image gpu_types.Image, requests []GPUImageWriteRequest) error {
	defer tracing.NewRegion("GPUDevice.WriteBufferToImageRegion").End()
	return g.writeBufferToImageRegionImpl(image, requests)
}
