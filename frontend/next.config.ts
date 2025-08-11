import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === "development";

const nextConfig: NextConfig = {
  /* 基本配置 */
  // 内存优化 - 在开发模式下禁用一些功能
  reactStrictMode: !isDev,
  productionBrowserSourceMaps: false,

  // 减少运行时代码
  poweredByHeader: false,
  
  // 图片配置 - 静态导出模式下禁用优化
  images: {
    unoptimized: true,
  },
  
  // 自定义webpack配置
  webpack: (config, { dev, isServer }) => {
    // 在开发模式下优化内存使用
    if (dev) {
      // 降低块大小
      config.optimization.splitChunks = {
        chunks: 'all',
        maxSize: 200000,
      };

      // 禁用性能提示，减少日志和内存使用
      config.performance = {
        hints: false,
        maxAssetSize: 1024 * 1024, // 1MB
        maxEntrypointSize: 1024 * 1024, // 1MB
      };
    }

    return config;
  },
  
  experimental: {
    // 优化编译器选项
    optimizePackageImports: ['@chakra-ui/react', 'react-icons'],
    
    // 开发过程中关闭预渲染以减少内存使用
    typedRoutes: !isDev,
  },

  // ESLint 配置
  eslint: {
    // 在生产构建时忽略 ESLint 错误
    ignoreDuringBuilds: true,
  },

  // TypeScript 配置
  typescript: {
    // 在生产构建时忽略类型检查错误
    ignoreBuildErrors: true,
  },

  // 启用静态导出，用于单体应用
  output: 'export',
  trailingSlash: true,
  skipTrailingSlashRedirect: true,
};

export default nextConfig;
