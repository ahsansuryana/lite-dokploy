const colors: Record<string, string> = {
  running: 'bg-green-600 text-green-100',
  done: 'bg-green-600 text-green-100',
  deploying: 'bg-blue-600 text-blue-100',
  failed: 'bg-red-600 text-red-100',
  stopped: 'bg-gray-600 text-gray-100',
  error: 'bg-red-600 text-red-100',
}

export function StatusBadge({ status }: { status: string }) {
  const cls = colors[status] || 'bg-gray-700 text-gray-200'
  return (
    <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${cls}`}>
      {status}
    </span>
  )
}
