export interface Dept {
  id: number;
  parentId: number;
  ancestors: string;
  name: string;
  orderNum: number;
  leader: number;
  phone: string;
  email: string;
  status: number;
  remark: string;
  createdAt: string;
}

export interface DeptTree {
  id: number;
  label: string;
  children?: DeptTree[];
}

export interface DeptUser {
  id: number;
  username: string;
  nickname: string;
}
