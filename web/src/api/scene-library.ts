import request from '../utils/request';

export interface SceneLibraryItem {
  id: string;
  name: string;
  location: string;
  time: string;
  description?: string;
  prompt?: string;
  image_url?: string;
  created_at: string;
  updated_at: string;
}

export interface ListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
}

export interface ListResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export const sceneLibraryAPI = {
  // 列出场景库中的场景
  list(params: ListParams) {
    return request.get<ListResponse<SceneLibraryItem>>('/scene-library', { params });
  },

  // 添加场景到库
  addToLibrary(sceneId: string) {
    return request.post<SceneLibraryItem>(`/scenes/${sceneId}/add-to-library`);
  },

  // 从库中删除场景
  removeFromLibrary(id: string) {
    return request.delete(`/scene-library/${id}`);
  },

  // 从库中应用场景到当前场景
  applyFromLibrary(sceneId: string, libraryItemId: string) {
    return request.put(`/scenes/${sceneId}/image-from-library`, {
      library_item_id: libraryItemId
    });
  },

  // 获取单个库场景
  get(libraryItemId: string) {
    return request.get<SceneLibraryItem>(`/scene-library/${libraryItemId}`);
  }
};