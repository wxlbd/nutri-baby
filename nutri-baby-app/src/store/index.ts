/**
 * Store 统一导出 - 临时禁用以排查循环依赖
 *
 * CRITICAL: 此文件已临时禁用
 * 所有页面应该直接从具体模块导入:
 * - import { xxx } from '@/store/user'
 * - import { xxx } from '@/store/baby'
 *
 * 不要使用: import { xxx } from '@/store'
 */

// 临时导出空对象,防止编译错误
export {}

console.warn('[Store] index.ts 已禁用,请使用直接导入: @/store/user, @/store/baby 等')

