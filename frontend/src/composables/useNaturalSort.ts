const cnNumMap: Record<string, number> = {
  '零': 0, '一': 1, '二': 2, '三': 3, '四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9,
  '十': 10, '廿': 20, '卅': 30, '卌': 40, '百': 100, '千': 1000, '万': 10000
}

const parseChineseNumber = (str: string): number => {
  let result = 0
  let temp = 0
  for (let i = 0; i < str.length; i++) {
    const char = str[i]
    if (cnNumMap[char]) {
      if (char === '十' && temp === 0) {
        temp = 1
      }
      const num = cnNumMap[char]
      if (num >= 10) {
        result += temp * num
        temp = 0
      } else {
        temp = temp * 10 + num
      }
    }
  }
  result += temp
  return result || NaN
}

const extractNumbers = (text: string): (string | number)[] => {
  const parts: (string | number)[] = []
  let currentText = ''
  
  const matches = text.matchAll(/(第?[\u4e00-\u9fa5]+[卷部回]?)|(\d+\.?\d*)/g)
  
  let lastIndex = 0
  for (const match of matches) {
    if (match.index! > lastIndex) {
      currentText += text.slice(lastIndex, match.index!)
    }
    if (currentText) {
      parts.push(currentText)
      currentText = ''
    }
    
    if (match[1]) {
      const cnMatch = match[1].match(/第?([\u4e00-\u9fa5]+)[卷部回]?/)
      if (cnMatch) {
        const num = parseChineseNumber(cnMatch[1])
        if (!isNaN(num)) {
          parts.push(num)
        } else {
          parts.push(match[1])
        }
      } else {
        parts.push(match[1])
      }
    } else if (match[2]) {
      parts.push(parseFloat(match[2]))
    }
    
    lastIndex = match.index! + match[0].length
  }
  
  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex))
  }
  
  if (parts.length === 0) {
    parts.push(text)
  }
  
  return parts
}

const compareParts = (a: string | number, b: string | number): number => {
  if (typeof a === 'number' && typeof b === 'number') {
    return a - b
  }
  if (typeof a === 'number') return -1
  if (typeof b === 'number') return 1
  return a.localeCompare(b, 'zh-CN')
}

export const naturalCompare = (a: string, b: string): number => {
  const aParts = extractNumbers(a.toLowerCase())
  const bParts = extractNumbers(b.toLowerCase())
  
  for (let i = 0; i < Math.max(aParts.length, bParts.length); i++) {
    const aPart = aParts[i] ?? ''
    const bPart = bParts[i] ?? ''
    const result = compareParts(aPart, bPart)
    if (result !== 0) return result
  }
  return 0
}