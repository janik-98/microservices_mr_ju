const url = window.location.href
let socket
let prices = []
let retryTimeout = 1000 // Initial retry timeout (in ms)

// Establish a WebSocket connection and set up event listeners
const connectWebSocket = () => {
  socket = new WebSocket(url.replace('http', 'ws'))

  // Listen for WebSocket open event
  socket.addEventListener('open', () => {
    console.log('WebSocket connected.')
    retryTimeout = 1000 // Reset the timeout after a successful connection
  })

  // Listen for messages from server
  socket.addEventListener('message', (event) => {
    const message = JSON.parse(event.data)
    switch (message.type) {
      case 'prices':
        prices = message.prices
        renderPrices()
        break
      default:
        console.error('Unknown message type:', message.type)
    }
  })

  // Listen for WebSocket close event and retry connection
  socket.addEventListener('close', () => {
    console.log('WebSocket closed. Reconnecting...')
    retryConnection()
  })

  // Listen for WebSocket errors
  socket.addEventListener('error', (event) => {
    console.error('WebSocket error:', event)
    socket.close() // Ensure the connection closes on error
  })
}

// Retry connection with exponential backoff
const retryConnection = () => {
  setTimeout(() => {
    console.log(`Retrying WebSocket connection in ${retryTimeout / 1000} seconds...`)
    retryTimeout = Math.min(retryTimeout * 2, 60000) // Cap timeout at 60 seconds
    connectWebSocket()
  }, retryTimeout)
}

// Render the prices
const renderPrices = () => {
  const pricesDiv = document.getElementById('prices')
  if (prices.length === 0) {
    return
  }
  pricesDiv.innerHTML = ''
  prices.forEach((price) => {
    const div = document.createElement('div')
    div.classList.add('flex', 'flex-col', 'items-center', 'justify-center', 'rounded', 'size-32', 'bg-slate-500')
    const h2 = document.createElement('h2')
    h2.classList.add('text-2xl', 'font-semibold')
    h2.innerText = price.company
    const span = document.createElement('span')
    span.classList.add('text-xl')
    span.innerText = price.avgPrice
    div.appendChild(h2)
    div.appendChild(span)
    pricesDiv.appendChild(div)
  })
}

// Initialize the WebSocket connection
connectWebSocket()
