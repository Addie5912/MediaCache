import fs from '@ohos.file.fs';
import zlib from '@ohos.zlib';
import util from '@ohos.util';
import common from '@ohos.app.ability.common';
import { BusinessError } from '@ohos.base';
export class LogUrlExtractor {
  private readonly LOG_DIR: string = '/data/log/';
  private context: common.Context;
  
  constructor(context: common.Context) {
    this.context = context;
  }
  
  async getLatestUrl(): Promise<string | null> {
    try {
      const latestGzFile = this.findLatestGzFile();
      if (!latestGzFile) {
        console.error('未找到符合条件的.gz文件');
        return null;
      }
      
      const decompressedPath = await this.decompressFile(latestGzFile);
      if (!decompressedPath) {
        return null;
      }
      
      const url = this.extractLatestUrl(decompressedPath);
      this.cleanupFile(decompressedPath);
      
      return url;
    } catch (error) {
      const err = error as BusinessError;
      console.error(`获取URL失败: ${err.message}`);
      return null;
    }
  }
  
  private findLatestGzFile(): string | null {
    try {
      const files = fs.listFileSync(this.LOG_DIR);
      const gzFiles: string[] = [];
      const pattern = /^log\d+\.gz$/;
      
      for (let i = 0; i < files.length; i++) {
        if (pattern.test(files[i])) {
          gzFiles.push(files[i]);
        }
      }
      
      if (gzFiles.length === 0) {
        return null;
      }
      
      gzFiles.sort((a, b) => b.localeCompare(a));
      return this.LOG_DIR + gzFiles[0];
    } catch (error) {
      const err = error as BusinessError;
      console.error(`读取目录失败: ${err.message}`);
      return null;
    }
  }
  
  private async decompressFile(gzFilePath: string): Promise<string | null> {
    try {
      const cacheDir = this.context.cacheDir;
      const fileName = gzFilePath.substring(gzFilePath.lastIndexOf('/') + 1);
      const decompressedFileName = fileName.replace('.gz', '');
      const decompressedPath = cacheDir + '/' + decompressedFileName;
      
      await zlib.decompressFile(gzFilePath, decompressedPath);
      
      return decompressedPath;
    } catch (error) {
      const err = error as BusinessError;
      console.error(`解压失败: ${err.message}, code: ${err.code}`);
      return null;
    }
  }
  
  private extractLatestUrl(filePath: string): string | null {
    try {
      const file = fs.openSync(filePath, fs.OpenMode.READ_ONLY);
      const stat = fs.statSync(filePath);
      const fileSize = stat.size;
      
      if (fileSize === 0) {
        fs.closeSync(file);
        return null;
      }
      
      const buffer = new ArrayBuffer(fileSize);
      fs.readSync(file.fd, buffer);
      fs.closeSync(file);
      
      const decoder = new util.TextDecoder('utf-8');
      const content = decoder.decodeToString(new Uint8Array(buffer));
      
      const urlPattern = /https:\/\/jdpull-ai\.jd\.com\/live\/[A-Za-z0-9]+_fhd\.flv/g;
      const matches = content.match(urlPattern);
      
      if (!matches || matches.length === 0) {
        return null;
      }
      
      return matches[matches.length - 1];
    } catch (error) {
      const err = error as BusinessError;
      console.error(`提取URL失败: ${err.message}`);
      return null;
    }
  }
  
  private cleanupFile(filePath: string): void {
    try {
      if (fs.accessSync(filePath)) {
        fs.unlinkSync(filePath);
      }
    } catch (error) {
      const err = error as BusinessError;
      console.error(`清理文件失败: ${err.message}`);
    }
  }
}
export async function getLatestPullUrl(context: common.Context): Promise<string | null> {
  const extractor = new LogUrlExtractor(context);
  return await extractor.getLatestUrl();
}


需要在 module.json5 中添加文件读写权限：
{
  "module": {
    "requestPermissions": [
      {
        "name": "ohos.permission.READ_MEDIA",
        "reason": "读取日志文件"
      },
      {
        "name": "ohos.permission.WRITE_MEDIA",
        "reason": "清理临时文件"
      }
    ]
  }
}


private findLatestGzFile(): string | null {
  try {
    // 先检查目录是否存在
    console.info(`检查目录: ${this.LOG_DIR}`);
    
    const isAccessible = this.checkDirAccess(this.LOG_DIR);
    if (!isAccessible) {
      console.error(`目录不可访问: ${this.LOG_DIR}`);
      console.error('可能原因: 1)目录不存在 2)权限不足 3)路径错误');
      return null;
    }
    
    const files = fs.listFileSync(this.LOG_DIR);
    console.info(`目录中的文件数量: ${files.length}`);
    
    const gzFiles: string[] = [];
    const pattern = /^log\d+\.gz$/;
    
    for (let i = 0; i < files.length; i++) {
      console.info(`发现文件: ${files[i]}`);
      if (pattern.test(files[i])) {
        gzFiles.push(files[i]);
      }
    }
    
    console.info(`匹配的.gz文件数量: ${gzFiles.length}`);
    
    if (gzFiles.length === 0) {
      console.error('未找到log*.gz文件');
      return null;
    }
    
    gzFiles.sort((a, b) => b.localeCompare(a));
    const latestFile = this.LOG_DIR + gzFiles[0];
    console.info(`最新文件: ${latestFile}`);
    
    return latestFile;
  } catch (error) {
    const err = error as BusinessError;
    console.error(`读取目录失败: code=${err.code}, message=${err.message}`);
    return null;
  }
}
/**
 * 检查目录是否可访问
 */
private checkDirAccess(dirPath: string): boolean {
  try {
    // 尝试访问目录
    fs.accessSync(dirPath);
    console.info(`目录存在且可访问: ${dirPath}`);
    return true;
  } catch (error) {
    const err = error as BusinessError;
    console.error(`目录访问失败: code=${err.code}, message=${err.message}`);
    return false;
  }
}
