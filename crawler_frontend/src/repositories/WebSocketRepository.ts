import { InvalidState } from "../common/InvalidState"

export interface ConnectedHandler {
    (): void
}

export interface DataHandler {
    (data: any): void
}

export interface WebSocketI {
    /// send will resolves once data is sent
    send(endpoint: string, data: any): Promise<void>
    close(): void

    registerConnectedHandler(handler: ConnectedHandler): void
    unregisterConnectedHandler(handler: ConnectedHandler): void
    
    registerDataHandler(handler: DataHandler): void
    unregisterDataHandler(handler: DataHandler): void
}

export class SimpleWebSocket implements WebSocketI {
    private dataHandlers: DataHandler[] = []
    private connectedHandlers: ConnectedHandler[] = []
    private isConnected = false
    private pendingDataToSend: any[] = []
    private messageCounter = 1
    constructor(
        private socket: WebSocket
    ) {
        this.socket.addEventListener("open",  this.connectionOpenHandler)
        if (this.socket.readyState === WebSocket.OPEN) {
            this.isConnected = true
        }
        this.socket.addEventListener("message", this.newDataHandler)
    }

    private connectionOpenHandler = () => {
        this.isConnected = true
        this.pendingDataToSend.forEach(value => {
            this.socket.send(value)
        })
        this.pendingDataToSend = []
        this.connectedHandlers.forEach(handler => handler())
    }

    private newDataHandler = (event: MessageEvent) => {
        let data = event.data
        try {
            data = JSON.parse(event.data)
        } catch (exception) {
            /// do nothing
        }
        this.dataHandlers.forEach(handler => handler(data))
    }

    send(endpoint: string, data: any): Promise<void> {
        let packagedData = {
            data: data,
            id: this.messageCounter + 1,
            type: endpoint,
        }
        if (this.socket.readyState !== this.socket.OPEN) {
            this.pendingDataToSend.push(JSON.stringify(packagedData))
        } else {
            this.pendingDataToSend = []
            this.socket.send(JSON.stringify(packagedData))
        }
        return new Promise((resolve) => resolve())
    }
    
    close(): void {
        this.isConnected = false
        this.socket.removeEventListener("open", this.connectionOpenHandler)
        this.socket.removeEventListener("message", this.newDataHandler)
    }

    registerConnectedHandler(handler: ConnectedHandler): void {
        if (this.isConnected) {
            handler()
        } else {
            this.connectedHandlers.push(handler)
        }
    }

    unregisterConnectedHandler(handler: ConnectedHandler): void {
        let index = this.connectedHandlers.indexOf(handler)
        if (index !== -1) {
            this.connectedHandlers.splice(index, 1)
        }
    }

    registerDataHandler(handler: DataHandler): void {
        this.dataHandlers.push(handler)
    }

    unregisterDataHandler(handler: DataHandler): void {
        let index = this.dataHandlers.indexOf(handler)
        if (index !== -1) {
            this.dataHandlers.splice(index, 1)
        }
    }
} 

export class WebSocketRepository {
    private socket: WebSocket;
    constructor(url: string) {
        this.socket = new WebSocket(url)
    }

    getNewSocket(): WebSocketI {
        return new SimpleWebSocket(this.socket)
    }
}