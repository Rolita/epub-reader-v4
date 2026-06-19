import JSZip from 'jszip'
import md5 from 'md5'

export interface BookMetadata {
  title: string
  author: string
  description: string // 简介
  coverData: Uint8Array | null
  coverPath: string
}

export interface BookConfig {
  id: string
  title: string
  author?: string
  description?: string // 简介
  originalFileName: string
  md5: string
  coverPath?: string
  createdAt: number
}

export interface ImportResult {
  success: boolean
  title: string
  author?: string
  coverUrl?: string
  md5: string
  filePath: string
  error?: string
}

/**
 * 将文件数据转换为 Uint8Array
 * Wails 在传输二进制数据时会进行 Base64 编码，需要解码
 */
export function toUint8Array(fileBytes: any): Uint8Array {
  if (fileBytes instanceof Uint8Array) {
    return fileBytes
  }
  
  if (Array.isArray(fileBytes)) {
    return new Uint8Array(fileBytes)
  }
  
  if (fileBytes && fileBytes.buffer) {
    return new Uint8Array(fileBytes.buffer)
  }
  
  if (typeof fileBytes === 'string') {
    // Base64 解码
    const binaryString = atob(fileBytes)
    const len = binaryString.length
    const buffer = new Uint8Array(len)
    for (let i = 0; i < len; i++) {
      buffer[i] = binaryString.charCodeAt(i)
    }
    return buffer
  }
  
  throw new Error('无法转换文件数据为 Uint8Array')
}

/**
 * 计算文件 MD5
 */
export function calculateMd5(buffer: Uint8Array): string {
  return md5(buffer)
}

/**
 * 从文件路径中提取文件名
 */
export function extractFileName(filePath: string): string {
  const parts = filePath.split(/[\\/]/)
  return parts[parts.length - 1]
}

/**
 * 从 EPUB 文件中提取元数据（标题、作者、封面、简介）
 * 采用多层兜底策略：OPF → 正则 → 文件名
 */
export async function extractEpubMetadata(fileBytes: Uint8Array): Promise<BookMetadata> {
  const zip = await JSZip.loadAsync(fileBytes)
  
  let coverData: Uint8Array | null = null
  let coverPath = ''
  let title = ''
  let author = ''
  let description = ''

  try {
    // 1. 找 OPF
    const containerXml = await zip.file('META-INF/container.xml')?.async('string')
    if (!containerXml) throw new Error('no container.xml')

    const opfPath = containerXml.match(/full-path="([^"]+)"/)?.[1]
    if (!opfPath) throw new Error('no opf path')

    const opfContent = await zip.file(opfPath)?.async('string')
    if (!opfContent) throw new Error('no opf content')

    // 2. 解析 XML（带命名空间）
    const xmlDoc = new DOMParser().parseFromString(opfContent, 'text/xml')
    const DC_NS = 'http://purl.org/dc/elements/1.1/'

    // —— 标题 ——
    const dcTitle = xmlDoc.getElementsByTagNameNS(DC_NS, 'title')[0]
    if (dcTitle) title = dcTitle.textContent?.trim() || ''
    if (!title) {
      const m = opfContent.match(/<dc:title\b[^>]*>([\s\S]*?)<\/dc:title>/i)
      if (m) title = m[1].trim()
    }

    // —— 作者（creator + contributor）——
    const authorList: string[] = []
    // creator
    const creators = xmlDoc.getElementsByTagNameNS(DC_NS, 'creator')
    for (let i = 0; i < creators.length; i++) {
      const t = creators[i].textContent?.trim()
      if (t) authorList.push(t)
    }
    // contributor（很多书把作者写在这里）
    const contributors = xmlDoc.getElementsByTagNameNS(DC_NS, 'contributor')
    for (let i = 0; i < contributors.length; i++) {
      const t = contributors[i].textContent?.trim()
      if (t && !authorList.includes(t)) authorList.push(t)
    }
    author = authorList.join(' / ')

    // 正则兜底抓 creator/contributor
    if (!author) {
      const reg1 = /<dc:creator\b[^>]*>([\s\S]*?)<\/dc:creator>/gi
      const reg2 = /<dc:contributor\b[^>]*>([\s\S]*?)<\/dc:contributor>/gi
      const list: string[] = []
      let m: RegExpExecArray | null
      while ((m = reg1.exec(opfContent))) list.push(m[1].trim())
      while ((m = reg2.exec(opfContent))) list.push(m[1].trim())
      author = list.join(' / ')
    }

    // —— 简介 description ——
    const dcDesc = xmlDoc.getElementsByTagNameNS(DC_NS, 'description')[0]
    if (dcDesc) description = dcDesc.textContent?.trim() || ''
    if (!description) {
      const m = opfContent.match(/<dc:description\b[^>]*>([\s\S]*?)<\/dc:description>/i)
      if (m) description = m[1].trim()
    }

    // —— 封面（沿用之前的稳健逻辑）——
    const opfDir = opfPath.substring(0, opfPath.lastIndexOf('/') + 1)
    let coverHref: string | null = null

    const items = xmlDoc.getElementsByTagName('item')
    for (let i = 0; i < items.length; i++) {
      const it = items[i]
      if (it.getAttribute('properties')?.includes('cover-image')) {
        coverHref = it.getAttribute('href')
        break
      }
    }
    if (!coverHref) {
      const metas = xmlDoc.getElementsByTagName('meta')
      let coverId = ''
      for (let i = 0; i < metas.length; i++) {
        if (metas[i].getAttribute('name') === 'cover') {
          coverId = metas[i].getAttribute('content') || ''
          break
        }
      }
      if (coverId) {
        for (let i = 0; i < items.length; i++) {
          if (items[i].getAttribute('id') === coverId) {
            coverHref = items[i].getAttribute('href')
            break
          }
        }
      }
    }
    if (coverHref) {
      const full = opfDir + coverHref
      const file = zip.file(full)
      if (file) {
        coverData = await file.async('uint8array')
        coverPath = full
      }
    }
    if (!coverData) {
      let maxSize = 0
      for (const [name, file] of Object.entries(zip.files)) {
        if (/\.(jpg|jpeg|png)$/i.test(name) && !file.dir) {
          const data = await file.async('uint8array')
          if (data.length > maxSize) {
            maxSize = data.length
            coverData = data
            coverPath = name
          }
        }
      }
    }
  } catch (e) {
    console.error('extract epub metadata error:', e)
  }

  return { title, author, description, coverData, coverPath }
}

/**
 * 保存书籍文件和相关资源
 */
async function saveBookFiles(
  bookDir: string,
  originalFileName: string,
  filePath: string,
  fileBytes: Uint8Array,
  metadata: BookMetadata,
  bookMd5: string
): Promise<{ coverUrl?: string; bookFilePath: string }> {
  // 保存封面
  let coverUrl: string | undefined
  if (metadata.coverData) {
    const coverArray = Array.from(metadata.coverData)
    // @ts-ignore
    await window.go.main.App.SaveFile(bookDir, 'cover.png', coverArray)
    coverUrl = `books/${bookMd5}/cover.png` // 相对路径，需要调用方补充书架名
  }

  // 复制书籍本体
  // @ts-ignore
  await window.go.main.App.CopyFile(filePath, bookDir, originalFileName)

  // 创建配置文件
  const config: BookConfig = {
    id: bookMd5,
    title: metadata.title,
    author: metadata.author,
    description: metadata.description || undefined,
    originalFileName,
    md5: bookMd5,
    coverPath: metadata.coverPath || undefined,
    createdAt: Date.now()
  }
  
  const configJson = JSON.stringify(config, null, 2)
  const configBytes = Array.from(new TextEncoder().encode(configJson))
  // @ts-ignore
  await window.go.main.App.SaveFile(bookDir, 'config.json', configBytes)

  const bookFilePath = `${bookDir}/${originalFileName}`
  
  return { coverUrl, bookFilePath }
}

/**
 * 导入书籍的完整流程
 * @param filePath 文件路径
 * @param shelfName 书架名称
 * @returns 导入结果
 */
export async function importBook(filePath: string, shelfName: string): Promise<ImportResult> {
  try {
    // 1. 读取文件内容
    // @ts-ignore
    const fileBytes = await window.go.main.App.GetFileBytes(filePath)
    
    console.log('原始 fileBytes 类型:', typeof fileBytes)
    console.log('原始 fileBytes 长度:', fileBytes?.length || fileBytes?.byteLength)
    
    // 2. 转换为 Uint8Array
    const buffer = toUint8Array(fileBytes)
    
    if (!buffer || buffer.byteLength < 1000) {
      console.error('警告：文件内容太小或为空，可能读取失败！')
      throw new Error('文件内容无效')
    }
    
    // 3. 计算 MD5
    const bookMd5 = calculateMd5(buffer)
    
    // 4. 获取书籍目录
    // @ts-ignore
    const booksDir = await window.go.main.App.GetBooksDir()
    const bookDir = `${booksDir}/${shelfName}/${bookMd5}`
    
    // 5. 提取文件名
    const originalFileName = extractFileName(filePath)
    
    // 6. 提取元数据
    const metadata = await extractEpubMetadata(buffer)
    
    // 7. 文件名兜底作者（关键！）
    if (!metadata.author) {
      const name = originalFileName.replace(/\.epub$/i, '')
      // 常见分隔符：—— _ - 【】 []
      const authorMatch = name.match(/^([^_\-——【\]]+)/)
      if (authorMatch) {
        metadata.author = authorMatch[1].trim()
      }
    }
    // 还空就显示未知
    if (!metadata.author) metadata.author = '未知作者'
    
    // 标题兜底
    if (!metadata.title) {
      metadata.title = originalFileName.replace(/\.epub$/i, '')
    }
    
    console.log('正在处理的文件:', originalFileName)
    console.log('文件字节数:', buffer.byteLength)
    console.log('提取的标题:', metadata.title)
    console.log('提取的作者:', metadata.author)
    
    // 8. 保存文件
    const { coverUrl, bookFilePath } = await saveBookFiles(
      bookDir,
      originalFileName,
      filePath,
      buffer,
      metadata,
      bookMd5
    )
    
    // 8. 构建完整的封面 URL
    const fullCoverUrl = coverUrl ? `books/${shelfName}/${bookMd5}/cover.png` : undefined
    
    return {
      success: true,
      title: metadata.title,
      author: metadata.author,
      coverUrl: fullCoverUrl,
      md5: bookMd5,
      filePath: bookFilePath
    }
  } catch (error) {
    console.error('导入书籍失败:', error)
    return {
      success: false,
      title: '',
      md5: '',
      filePath: '',
      error: error instanceof Error ? error.message : '未知错误'
    }
  }
}

/**
 * 批量导入书籍
 * @param filePaths 文件路径列表
 * @param shelfName 书架名称
 * @param onProgress 进度回调
 * @returns 导入结果列表
 */
export async function importBooks(
  filePaths: string[],
  shelfName: string,
  onProgress?: (current: number, total: number, title: string) => void
): Promise<ImportResult[]> {
  const results: ImportResult[] = []
  
  for (let i = 0; i < filePaths.length; i++) {
    const filePath = filePaths[i]
    
    if (onProgress) {
      onProgress(i + 1, filePaths.length, filePath.split(/[\\/]/).pop() || '')
    }
    
    const result = await importBook(filePath, shelfName)
    results.push(result)
  }
  
  return results
}

/**
 * 从文件夹导入书籍
 * 先扫描文件夹中的所有 EPUB 文件，然后批量导入
 * @param folderPath 文件夹路径
 * @param shelfName 书架名称
 * @param onProgress 进度回调
 * @returns 导入结果列表
 */
export async function importBooksFromFolder(
  folderPath: string,
  shelfName: string,
  onProgress?: (current: number, total: number, title: string) => void
): Promise<ImportResult[]> {
  // 扫描文件夹中的 EPUB 文件
  // @ts-ignore
  const filePaths: string[] = await window.go.main.App.ScanEpubFilesInFolder(folderPath)
  
  if (filePaths.length === 0) {
    return [{
      success: false,
      title: '',
      md5: '',
      filePath: '',
      error: '未找到 EPUB 文件'
    }]
  }
  
  return await importBooks(filePaths, shelfName, onProgress)
}