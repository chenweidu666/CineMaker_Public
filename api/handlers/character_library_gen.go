package handlers

// DEPRECATED: 注释掉角色图片生成handler - 目前只使用首帧图片生成
// func (h *CharacterLibraryHandler) GenerateCharacterImage(c *gin.Context) {
//
// 	characterID := c.Param("id")
//
// 	// 获取请求体中的model、style和prompt参数
// 	var req struct {
// 		Model  string `json:"model"`
// 		Style  string `json:"style"`
// 		Prompt string `json:"prompt"`
// 	}
// 	c.ShouldBindJSON(&req)
//
// 	imageGen, prompt, err := h.libraryService.GenerateCharacterImage(characterID, h.imageService, req.Model, req.Style, req.Prompt)
// 	if err != nil {
// 		if err.Error() == "character not found" {
// 			response.NotFound(c, "角色不存在")
// 			return
// 		}
// 		if err.Error() == "unauthorized" {
// 			response.Forbidden(c, "无权限")
// 			return
// 		}
// 		h.log.Errorw("Failed to generate character image", "error", err)
// 		response.InternalError(c, "生成失败")
// 		return
// 	}
//
// 	response.Success(c, gin.H{
// 		"message":          "角色图片生成已启动",
// 		"image_generation": imageGen,
// 		"prompt":           prompt,
// 	})
// }
