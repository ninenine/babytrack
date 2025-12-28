import { useState } from 'react'
import { AlertTriangle, Copy, Check, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface ErrorFallbackProps {
  error: Error
  errorInfo: React.ErrorInfo | null
  onReset: () => void
}

export function ErrorFallback({ error, errorInfo, onReset }: ErrorFallbackProps) {
  const [copied, setCopied] = useState(false)

  const errorDetails = `Error: ${error.message}

Stack Trace:
${error.stack || 'No stack trace available'}

Component Stack:
${errorInfo?.componentStack || 'No component stack available'}

Timestamp: ${new Date().toISOString()}
URL: ${window.location.href}
User Agent: ${navigator.userAgent}`

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(errorDetails)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Clipboard API failed - likely permissions issue
      console.error('Failed to copy to clipboard')
    }
  }

  const handleReload = () => {
    window.location.reload()
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <div className="max-w-2xl w-full space-y-6">
        <div className="flex flex-col items-center text-center space-y-4">
          <div className="rounded-full bg-destructive/10 p-4">
            <AlertTriangle className="h-12 w-12 text-destructive" />
          </div>
          <div className="space-y-2">
            <h1 className="text-2xl font-bold">Something went wrong</h1>
            <p className="text-muted-foreground">
              An unexpected error occurred. You can try refreshing the page or copy the error
              details to report the issue.
            </p>
          </div>
        </div>

        <div className="rounded-lg border bg-muted/50 p-4 space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Error Details</span>
            <Button variant="outline" size="sm" onClick={handleCopy} className="gap-2">
              {copied ? (
                <>
                  <Check className="h-4 w-4" />
                  Copied
                </>
              ) : (
                <>
                  <Copy className="h-4 w-4" />
                  Copy Error
                </>
              )}
            </Button>
          </div>
          <div className="rounded-md bg-background border p-3 max-h-64 overflow-auto">
            <pre className="text-xs text-muted-foreground whitespace-pre-wrap wrap-break-word font-mono">
              {error.message}
              {error.stack && (
                <>
                  {'\n\n'}
                  {error.stack}
                </>
              )}
            </pre>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <Button onClick={onReset} variant="outline" className="gap-2">
            <RefreshCw className="h-4 w-4" />
            Try Again
          </Button>
          <Button onClick={handleReload} className="gap-2">
            <RefreshCw className="h-4 w-4" />
            Reload Page
          </Button>
        </div>
      </div>
    </div>
  )
}
