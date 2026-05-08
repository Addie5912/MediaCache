import fs from '@ohos.file.fs';
import zlib from '@ohos.zlib';
import util from '@ohos.util';
import common from '@ohos.app.ability.common';
export class LogUrlExtractor {
  private readonly LOG_DIR: string = '/data/log/';  // 只读的源目录
  private context: common.Context;  // 应用上下文，用于获取缓存目录
  
  constructor(context: common.Context) {
    this.context = context;
  }
  
  /**
   * 主入口：获取最新的jdpull-ai.jd.com URL
   */
  async getLatestUrl(): Promise<string | null> {
    try {
      // 1. 获取最新的.gz文件
      const latestGzFile = this.findLatestGzFile();
      if (!latestGzFile) {
        console.error('未找到符合条件的.gz文件');
        return null;
      }
      
      console.info(`找到最新文件: ${latestGzFile}`);
      
      // 2. 解压文件到应用缓存目录（有写权限）
      const decompressedPath = await this.decompressFile(latestGzFile);
      if (!decompressedPath) {
        console.error('解压失败');
        return null;
      }
      
      // 3. 读取解压后的内容并过滤URL
      const url = this.extractLatestUrl(decompressedPath);
      
      // 4. 清理临时文件
      this.cleanupFile(decompressedPath);
      
      return url;
    } catch (error) {
      console.error(`获取URL失败: ${JSON.stringify(error)}`);
      return null;
    }
  }
  
  /**
   * 查找最新的log*.gz文件（只读操作）
   */
  private findLatestGzFile(): string | null {
    try {
      // 读取目录内容
      const files = fs.listFileSync(this.LOG_DIR);
      
      // 过滤出log*.gz文件
      const gzFiles: string[] = [];
      const pattern = /^log\d+\.gz$/;
      
      for (let i = 0; i < files.length; i++) {
        const fileName = files[i];
        if (pattern.test(fileName)) {
          gzFiles.push(fileName);
        }
      }
      
      if (gzFiles.length === 0) {
        console.error('未找到log*.gz文件');
        return null;
      }
      
      // 按文件名排序（文件名包含时间戳，字符串排序即可获取最新）
      gzFiles.sort((a, b) => b.localeCompare(a));
      
      // 返回最新文件的完整路径
      return this.LOG_DIR + gzFiles[0];
    } catch (error) {
      console.error(`读取目录失败: ${JSON.stringify(error)}`);
      return null;
    }
  }
  
  /**
   * 解压.gz文件到应用缓存目录
   * @param gzFilePath .gz文件路径（只读）
   * @returns 解压后的文件路径（应用缓存目录）
   */
  private async decompressFile(gzFilePath: string): Promise<string | null> {
    try {
      // 获取应用缓存目录（有写权限）
      const cacheDir = this.context.cacheDir;
      
      // 生成解压后的文件路径（在应用缓存目录中）
      const fileName = gzFilePath.substring(gzFilePath.lastIndexOf('/') + 1);
      const decompressedFileName = fileName.replace('.gz', '');
      const decompressedPath = cacheDir + '/' + decompressedFileName;
      
      console.info(`准备解压到: ${decompressedPath}`);
      
      // 使用zlib解压
      await zlib.decompressFile(gzFilePath, decompressedPath, (errCode: number) => {
        if (errCode !== 0) {
          console.error(`解压失败，错误码: ${errCode}`);
          return null;
        }
      });
      
      console.info(`解压成功: ${decompressedPath}`);
      return decompressedPath;
    } catch (error) {
      console.error(`解压过程异常: ${JSON.stringify(error)}`);
      return null;
    }
  }
  
  /**
   * 从解压文件中提取最新的jdpull-ai.jd.com URL
   * @param filePath 解压后的文件路径
   * @returns 最新的URL
   */
  private extractLatestUrl(filePath: string): string | null {
    try {
      // 读取文件内容
      const file = fs.openSync(filePath, fs.OpenMode.READ_ONLY);
      const stat = fs.statSync(filePath);
      const fileSize = stat.size;
      
      if (fileSize === 0) {
        console.error('文件为空');
        fs.closeSync(file);
        return null;
      }
      
      // 创建缓冲区读取内容
      const buffer = new ArrayBuffer(fileSize);
      fs.readSync(file.fd, buffer);
      fs.closeSync(file);
      
      // 转换为字符串
      const decoder = new util.TextDecoder('utf-8');
      const content = decoder.decodeToString(new Uint8Array(buffer));
      
      // 正则匹配URL
      const urlPattern = /https:\/\/jdpull-ai\.jd\.com\/live\/[A-Za-z0-9]+_fhd\.flv/g;
      const matches = content.match(urlPattern);
      
      if (!matches || matches.length === 0) {
        console.error('未找到匹配的URL');
        return null;
      }
      
      // 返回最后一个（最新的）URL
      const latestUrl = matches[matches.length - 1];
      console.info(`找到最新URL: ${latestUrl}`);
      
      return latestUrl;
    } catch (error) {
      console.error(`提取URL失败: ${JSON.stringify(error)}`);
      return null;
    }
  }
  
  /**
   * 清理临时文件
   */
  private cleanupFile(filePath: string): void {
    try {
      if (fs.accessSync(filePath)) {
        fs.unlinkSync(filePath);
        console.info(`已清理临时文件: ${filePath}`);
      }
    } catch (error) {
      console.error(`清理文件失败: ${JSON.stringify(error)}`);
    }
  }
}
// 使用示例
export async function getLatestPullUrl(context: common.Context): Promise<string | null> {
  const extractor = new LogUrlExtractor(context);
  return await extractor.getLatestUrl();
}
