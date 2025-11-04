import { useState, useEffect, useRef } from 'react'
import { Select } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Search, Download, Trash2, Pause, Play, ArrowDown } from 'lucide-react'
import Convert from 'ansi-to-html'

const ansiConverter = new Convert({
  fg: '#FFF',
  bg: '#000',
  newline: false,
  escapeXML: true,
  stream: false
})

interface LogEntry {
  service: string
  message: string
  level: number
  timestamp: string
  isStderr: boolean
}

export function LogsView() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [services, setServices] = useState<string[]>([])
  const [selectedService, setSelectedService] = useState<string>('all')
  const [searchTerm, setSearchTerm] = useState('')
  const [isPaused, setIsPaused] = useState(false)
  const [isUserScrolling, setIsUserScrolling] = useState(false)
  const logsEndRef = useRef<HTMLDivElement>(null)
  const logsContainerRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)

  // Fetch services list
  useEffect(() => {
    fetch('/api/services')
      .then(res => res.json())
      .then(data => {
        const serviceNames = data.map((s: any) => s.name)
        setServices(serviceNames)
      })
      .catch(err => console.error('Failed to fetch services:', err))
  }, [])

  // Fetch initial logs and setup WebSocket
  useEffect(() => {
    fetchLogs()
    setupWebSocket()

    return () => {
      wsRef.current?.close()
    }
  }, [selectedService])

  // Auto-scroll to bottom
  useEffect(() => {
    if (!isPaused && !isUserScrolling) {
      logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [logs, isPaused, isUserScrolling])

  // Detect manual scrolling
  const handleScroll = () => {
    const container = logsContainerRef.current
    if (!container) return

    const { scrollTop, scrollHeight, clientHeight } = container
    const isAtBottom = Math.abs(scrollHeight - clientHeight - scrollTop) < 10

    if (!isAtBottom && !isUserScrolling) {
      // User scrolled up - pause auto-scroll
      setIsUserScrolling(true)
      setIsPaused(true)
    } else if (isAtBottom && isUserScrolling) {
      // User scrolled back to bottom - resume auto-scroll
      setIsUserScrolling(false)
      setIsPaused(false)
    }
  }

  const fetchLogs = async () => {
    const url = selectedService === 'all'
      ? '/api/logs?tail=500'
      : `/api/logs?service=${selectedService}&tail=500`

    console.log('Fetching initial logs from:', url)
    try {
      const res = await fetch(url)
      const data = await res.json()
      console.log('Fetched logs:', data?.length, 'entries', data)
      setLogs(data || [])
    } catch (err) {
      console.error('Failed to fetch logs:', err)
      setLogs([])
    }
  }

  const setupWebSocket = () => {
    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close()
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = selectedService === 'all'
      ? `${protocol}//${window.location.host}/api/logs/stream`
      : `${protocol}//${window.location.host}/api/logs/stream?service=${selectedService}`

    console.log('Setting up WebSocket connection to:', url)
    const ws = new WebSocket(url)

    ws.onopen = () => {
      console.log('WebSocket connected successfully')
    }

    ws.onmessage = (event) => {
      console.log('Received log entry:', event.data)
      if (!isPaused) {
        try {
          const entry = JSON.parse(event.data)
          setLogs(prev => [...prev, entry].slice(-1000)) // Keep last 1000
        } catch (err) {
          console.error('Failed to parse log entry:', err)
        }
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    ws.onclose = () => {
      console.log('WebSocket closed')
    }

    wsRef.current = ws
  }

  const filteredLogs = logs.filter(log =>
    log && log.message && log.message.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const exportLogs = () => {
    const content = filteredLogs
      .map(log => `[${log.timestamp || ''}] [${log.service || ''}] ${log.message || ''}`)
      .join('\n')

    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-${Date.now()}.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  const clearLogs = () => {
    if (window.confirm(`Clear all ${logs.length} log entries? This cannot be undone.`)) {
      setLogs([])
    }
  }

  const togglePause = () => {
    const newPausedState = !isPaused
    setIsPaused(newPausedState)
    
    // If resuming, scroll to bottom
    if (!newPausedState) {
      setIsUserScrolling(false)
      setTimeout(() => {
        logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
      }, 100)
    }
  }

  const scrollToBottom = () => {
    setIsUserScrolling(false)
    setIsPaused(false)
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp)
      const time = date.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
      const ms = date.getMilliseconds().toString().padStart(3, '0')
      return `${time}.${ms}`
    } catch {
      return timestamp
    }
  }

  // Error/warning detection regex
  const isErrorLine = (message: string) => {
    const errorPattern = /\b(error|failed|failure|exception|fatal|panic|critical|crash|died)\b/i
    return errorPattern.test(message)
  }

  const isWarningLine = (message: string) => {
    const warningPattern = /\b(warn|warning|caution|deprecated)\b/i
    return warningPattern.test(message)
  }

  // Assign consistent colors to services (avoiding red)
  const serviceColors = [
    'text-blue-400',
    'text-green-400', 
    'text-purple-400',
    'text-cyan-400',
    'text-pink-400',
    'text-amber-400',
    'text-teal-400',
    'text-indigo-400',
    'text-lime-400',
    'text-fuchsia-400',
    'text-sky-400',
    'text-violet-400',
  ]

  const getServiceColor = (serviceName: string) => {
    // Generate consistent color index from service name
    const hash = serviceName.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
    return serviceColors[hash % serviceColors.length]
  }

  const getLogColor = (log: LogEntry) => {
    // Check message content first for errors/warnings
    if (isErrorLine(log.message)) return 'text-red-400'
    if (isWarningLine(log.message)) return 'text-yellow-400'
    
    // Check log level and stderr
    if (log.isStderr || log.level === 3) return 'text-red-400'
    if (log.level === 2) return 'text-yellow-400'
    if (log.level === 1) return 'text-gray-400'
    
    return 'text-foreground'
  }

  const convertAnsiToHtml = (text: string) => {
    try {
      return ansiConverter.toHtml(text)
    } catch {
      // If conversion fails, return original text
      return text
    }
  }

  return (
    <div className="space-y-4">
      {/* Controls */}
      <div className="flex gap-4 items-center flex-wrap">
        <Select 
          value={selectedService} 
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setSelectedService(e.target.value)}
          className="min-w-[150px]"
        >
          <option value="all">All Services</option>
          {services.map((service) => (
            <option key={service} value={service}>{service}</option>
          ))}
        </Select>

        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-3 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search logs..."
            value={searchTerm}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>

        <Button
          variant="outline"
          size="icon"
          onClick={togglePause}
          title={isPaused ? 'Resume' : 'Pause'}
        >
          {isPaused ? <Play className="w-4 h-4" /> : <Pause className="w-4 h-4" />}
        </Button>

        <Button variant="outline" size="icon" onClick={exportLogs} title="Export logs">
          <Download className="w-4 h-4" />
        </Button>

        <Button variant="outline" size="icon" onClick={clearLogs} title="Clear logs">
          <Trash2 className="w-4 h-4" />
        </Button>
      </div>

      {/* Log Display */}
      <div 
        ref={logsContainerRef}
        onScroll={handleScroll}
        className="bg-card border rounded-lg p-4 h-[600px] overflow-y-auto font-mono text-sm"
      >
        {filteredLogs.length === 0 ? (
          <div className="text-center text-muted-foreground py-12">
            {logs.length === 0 ? 'No logs to display' : 'No logs match your search'}
          </div>
        ) : (
          <div className="space-y-0.5">
            {filteredLogs.map((log, idx) => (
              <div key={idx} className={getLogColor(log)}>
                <span className="text-muted-foreground text-xs">
                  [{formatTimestamp(log?.timestamp || '')}]
                </span>
                {' '}
                <span className={getServiceColor(log?.service || 'unknown')}>
                  [{log?.service || 'unknown'}]
                </span>
                {' '}
                <span 
                  dangerouslySetInnerHTML={{ 
                    __html: convertAnsiToHtml(log?.message || '') 
                  }} 
                />
              </div>
            ))}
            <div ref={logsEndRef} />
          </div>
        )}
      </div>

      {/* Status Bar */}
      <div className="text-sm text-muted-foreground flex justify-between items-center">
        <span>
          Showing {filteredLogs.length} of {logs.length} log entries
        </span>
        <div className="flex items-center gap-4">
          {isPaused && (
            <>
              <span className="text-yellow-600 font-medium">⏸ Paused - scroll stopped</span>
              <Button 
                variant="outline" 
                size="sm" 
                onClick={scrollToBottom}
                className="flex items-center gap-2"
              >
                <ArrowDown className="w-4 h-4" />
                Jump to Bottom
              </Button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
