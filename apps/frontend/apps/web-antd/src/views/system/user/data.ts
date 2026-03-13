import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'username',
    label: '用户账号',
  },
  {
    component: 'Input',
    fieldName: 'nickname',
    label: '用户昵称',
  },
  {
    component: 'Input',
    fieldName: 'phone',
    label: '手机号码',
  },
  {
    component: 'Select',
    fieldName: 'status',
    label: '用户状态',
    componentProps: {
      options: [
        { label: '启用', value: 1 },
        { label: '禁用', value: 0 },
      ],
    },
  },
  {
    component: 'RangePicker',
    fieldName: 'createdAt',
    label: '创建时间',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    field: 'username',
    title: '名称',
    minWidth: 120,
    sortable: true,
  },
  {
    field: 'nickname',
    title: '昵称',
    minWidth: 120,
    sortable: true,
  },
  {
    field: 'phone',
    title: '手机号',
    formatter({ cellValue }) {
      return cellValue || '暂无';
    },
    minWidth: 130,
    sortable: true,
  },
  {
    field: 'email',
    title: '邮箱',
    minWidth: 160,
    sortable: true,
  },
  {
    field: 'status',
    title: '状态',
    minWidth: 100,
    slots: { default: 'status' },
    sortable: true,
  },
  {
    field: 'createdAt',
    title: '创建时间',
    minWidth: 180,
    sortable: true,
  },
  {
    field: 'action',
    slots: { default: 'action' },
    title: '操作',
    fixed: 'right',
    resizable: false,
    width: 'auto',
  },
];

/** 新增/编辑表单schema */
export function drawerSchema(isEdit: boolean): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: '用户名',
      rules: 'required',
      componentProps: {
        placeholder: '请输入用户名',
        disabled: isEdit,
      },
    },
    {
      component: 'InputPassword',
      fieldName: 'password',
      label: '密码',
      rules: isEdit ? undefined : 'required',
      componentProps: {
        placeholder: isEdit ? '留空则不修改' : '请输入密码',
      },
    },
    {
      component: 'Input',
      fieldName: 'nickname',
      label: '昵称',
      componentProps: {
        placeholder: '请输入昵称',
      },
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: '邮箱',
      componentProps: {
        placeholder: '请输入邮箱',
      },
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: '手机号',
      componentProps: {
        placeholder: '请输入手机号',
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'status',
      label: '状态',
      defaultValue: 1,
      componentProps: {
        options: [
          { label: '正常', value: 1 },
          { label: '停用', value: 0 },
        ],
      },
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: '备注',
      componentProps: {
        placeholder: '请输入备注',
        rows: 3,
      },
    },
  ];
}
