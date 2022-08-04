import { Item } from "./Item"

export type PriceRequestId = number

export interface PriceRequest {
    id: number
    items: Item[]
}