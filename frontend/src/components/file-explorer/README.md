# File Explorer UI Components

一个基于 Chakra UI 和 React Dropzone 构建的完整文件管理器 UI 组件库。

## 功能特性

- ✅ **BreadcrumbNav**: 面包屑导航，显示当前路径并支持快速跳转
- ✅ **ObjectTable**: 文件/文件夹列表，支持排序、多选、响应式布局
- ✅ **UploadDropzone**: 拖拽上传区域，支持进度显示和并发上传
- ✅ **ContextMenu**: 右键菜单，支持重命名、删除、下载等操作
- ✅ **Modal 对话框**: 重命名和新建文件夹的模态框
- ✅ **响应式设计**: 移动端自动切换到卡片列表视图
- ✅ **文件夹优先显示**
- ✅ **双击导航和下载**
- ✅ **并发上传进度条**

## 组件结构

```
src/components/file-explorer/
├── FileExplorer.tsx          # 主组件，组合所有子组件
├── BreadcrumbNav.tsx         # 面包屑导航
├── ObjectTable.tsx           # 文件/文件夹表格
├── UploadDropzone.tsx        # 上传拖拽区域
├── ContextMenu.tsx           # 右键菜单
├── RenameModal.tsx          # 重命名对话框
├── NewFolderModal.tsx       # 新建文件夹对话框
├── index.ts                 # 导出文件
└── README.md               # 文档
```

## 使用方法

### 基本使用

```typescript
import { FileExplorer, FileSystemItem, FileExplorerActions } from '@/components/file-explorer';

function MyFilePage() {
  const [currentPath, setCurrentPath] = useState('/');
  const [items, setItems] = useState<FileSystemItem[]>([]);
  const [uploads, setUploads] = useState<UploadProgress[]>([]);

  const actions: FileExplorerActions = {
    onNavigate: (path) => setCurrentPath(path),
    onFileDoubleClick: (file) => downloadFile(file),
    onFolderDoubleClick: (folder) => setCurrentPath(folder.path),
    onUpload: (files, path) => uploadFiles(files, path),
    onRename: (item, newName) => renameItem(item, newName),
    onDelete: (items) => deleteItems(items),
    onDownload: (items) => downloadItems(items),
    onCreateFolder: (path, name) => createFolder(path, name),
    onRefresh: () => loadItems(currentPath),
  };

  return (
    <FileExplorer
      currentPath={currentPath}
      items={items}
      breadcrumbs={generateBreadcrumbs(currentPath)}
      uploads={uploads}
      actions={actions}
      config={{
        allowMultiSelect: true,
        maxFileSize: 100 * 1024 * 1024, // 100MB
        enableDragAndDrop: true,
        enableContextMenu: true,
      }}
    />
  );
}
```

### 单独使用子组件

```typescript
import { ObjectTable, UploadDropzone, BreadcrumbNav } from '@/components/file-explorer';

// 只使用文件表格
<ObjectTable
  items={files}
  selectedItems={selectedIds}
  onItemSelect={(id, selected) => handleSelect(id, selected)}
  onItemDoubleClick={(item) => handleDoubleClick(item)}
  allowMultiSelect={true}
/>

// 只使用上传组件
<UploadDropzone
  currentPath="/documents"
  uploads={uploadProgress}
  onUpload={(files) => handleUpload(files)}
  maxSize={50 * 1024 * 1024}
  multiple={true}
/>

// 只使用面包屑导航
<BreadcrumbNav
  items={breadcrumbs}
  onNavigate={(path) => setCurrentPath(path)}
  onRefresh={() => refresh()}
/>
```

## 类型定义

### FileSystemItem

```typescript
interface FileSystemItem {
  id: string;
  name: string;
  type: 'file' | 'folder';
  size: number;
  lastModified: Date;
  path: string;
  parentPath?: string;
  mimeType?: string;
  // ... 其他属性
}
```

### FileExplorerActions

```typescript
interface FileExplorerActions {
  onNavigate: (path: string) => void;
  onFileSelect: (file: FileSystemItem) => void;
  onFileDoubleClick: (file: FileSystemItem) => void;
  onFolderDoubleClick: (folder: FileSystemItem) => void;
  onUpload: (files: File[], path: string) => void;
  onRename: (item: FileSystemItem, newName: string) => void;
  onDelete: (items: FileSystemItem[]) => void;
  onDownload: (items: FileSystemItem[]) => void;
  onCreateFolder: (path: string, name: string) => void;
  onRefresh: () => void;
}
```

### FileExplorerConfig

```typescript
interface FileExplorerConfig {
  allowMultiSelect?: boolean;        // 是否允许多选
  showHiddenFiles?: boolean;         // 是否显示隐藏文件
  allowedFileTypes?: string[];       // 允许的文件类型
  maxFileSize?: number;              // 最大文件大小（字节）
  maxConcurrentUploads?: number;     // 最大并发上传数
  enableDragAndDrop?: boolean;       // 是否启用拖拽
  enableContextMenu?: boolean;       // 是否启用右键菜单
}
```

## 特殊功能说明

### 1. 响应式设计

组件会根据屏幕尺寸自动切换显示模式：
- **桌面端**: 表格视图，显示完整的列信息
- **移动端**: 卡片视图，隐藏部分列以适应小屏幕

### 2. 文件图标

根据文件类型和扩展名自动显示对应图标：
- 文件夹: 蓝色文件夹图标
- 图片: 图片图标 (.jpg, .png, .gif 等)
- 视频: 视频图标 (.mp4, .avi, .mkv 等)
- 音频: 音频图标 (.mp3, .wav, .flac 等)
- 压缩包: 压缩包图标 (.zip, .rar, .7z 等)
- 文档: 文档图标 (.pdf, .doc, .txt 等)
- 代码: 代码图标 (.js, .ts, .html 等)

### 3. 拖拽上传

支持以下上传方式：
- 拖拽文件到上传区域
- 点击选择文件按钮
- 工具栏中的上传按钮

### 4. 上传进度

- 显示每个文件的上传进度
- 支持取消和重试
- 显示上传统计信息
- 自动清理已完成的上传

### 5. 右键菜单

根据选中项目类型和数量动态显示菜单项：
- 单个文件: 预览、下载、重命名、复制、移动、分享、属性、删除
- 单个文件夹: 下载、重命名、复制、移动、分享、属性、删除
- 多个项目: 下载、复制、移动、分享、删除
- 空白区域: 新建文件夹、刷新

### 6. 键盘支持

- **Enter**: 打开选中项目
- **Delete**: 删除选中项目
- **F2**: 重命名选中项目
- **Ctrl+A**: 全选
- **Escape**: 取消选择

## 自定义样式

组件使用 Chakra UI 的主题系统，可以通过主题配置来自定义样式：

```typescript
import { extendTheme } from '@chakra-ui/react';

const theme = extendTheme({
  components: {
    Button: {
      defaultProps: {
        colorScheme: 'blue', // 默认按钮颜色
      },
    },
  },
});
```

## 示例

完整的使用示例请参考 `src/app/file-explorer/page.tsx`。

## 依赖

- React 18+
- Chakra UI 2.8+
- React Dropzone 14+
- Framer Motion 11+
- React Icons 5+

## 浏览器支持

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+
