import request from '../utils/request';

export interface PropLibraryItem {
  id: number;
  name: string;
  type?: string;
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

export const propLibraryAPI = {
  // 列出道具库中的道具
  list(params: ListParams) {
    return request.get<ListResponse<PropLibraryItem>>('/prop-library', { params });
  },

  // 添加道具到库
  addToLibrary(propId: number) {
    return request.post<PropLibraryItem>(`/props/${propId}/add-to-library`);
  },

  // 从库中删除道具
  removeFromLibrary(id: number) {
    return request.delete(`/prop-library/${id}`);
  },

  // 从库中应用道具到当前道具
  applyFromLibrary(propId: number, libraryItemId: number) {
    return request.put(`/props/${propId}/image-from-library`, {
      library_item_id: libraryItemId
    });
  },

  // 获取单个库道具
  get(libraryItemId: number) {
    return request.get<PropLibraryItem>(`/prop-library/${libraryItemId}`);
  }
};