import { ResourceBusy } from "../common/ResourceBusy";
import { Item } from "../models/Item";
import { PriceRequest, PriceRequestId } from "../models/PriceRequest";
import { PriceRequestRepositoryI } from "./PriceRequestRepositoryI";
import { WebSocketI } from "./WebSocketRepository";

type RemotePriceRequestId = number
interface PriceRequestData {
    type: string,
    data: PriceRequest,
}

const PriceRequestDataType = "price-request"

function isPriceRequestData(data: any): data is PriceRequestData {
    return 'type' in data && 'data' in data && (data.type === PriceRequestDataType)
}

export class PriceRequestRepositoryWs implements PriceRequestRepositoryI {
    private resolve: ((value: PriceRequest) => void) | null = null
    constructor(
        private socket: WebSocketI
    ) {
        this.socket.registerDataHandler(this.dataHandler)
    }

    close() {
        this.socket.unregisterDataHandler(this.dataHandler)
    }

    private dataHandler = (data: any) => {
        if (isPriceRequestData(data)) {
            if (data.type === PriceRequestDataType) {
                /// handle
                if (this.resolve !== null) {
                    let priceRequest = data.data
                    let resolve = this.resolve
                    this.resolve = null
                    resolve(priceRequest)
                }
            } else {
            }
        }
    }

    async postNewRequest(keyword: string): Promise<PriceRequest> {
        if (this.resolve !== null) {
            /// throw ResourceBusy
            throw new ResourceBusy("cannot send new request as another is ongoing. Cancel previous request first")
        }  

        let request = {
            keyword: keyword,
        }
        
        console.log('send')
        await this.socket.send("price-request", request)
        return new Promise((resolve, reject) => {
            this.resolve = resolve
        })
    }

    fetchRequestResults(id: number): Promise<Item[] | null> {
        throw new Error("Method not implemented.");
    }

}