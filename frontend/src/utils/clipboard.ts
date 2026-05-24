/**
 * Copy text to clipboard with fallback for non-secure contexts (HTTP).
 * In HTTPS or localhost, uses the modern Clipboard API.
 * In HTTP, falls back to document.execCommand('copy').
 */
export async function copyToClipboard(text: string): Promise<void> {
  // Secure context: use modern Clipboard API
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text)
  }

  // Fallback for HTTP (non-secure context)
  const textArea = document.createElement('textarea')
  textArea.value = text
  textArea.style.position = 'fixed'
  textArea.style.left = '-9999px'
  textArea.style.top = '-9999px'
  document.body.appendChild(textArea)
  textArea.focus()
  textArea.select()

  try {
    document.execCommand('copy')
  } finally {
    document.body.removeChild(textArea)
  }
}
