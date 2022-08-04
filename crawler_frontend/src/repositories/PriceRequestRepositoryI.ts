import { Item } from "../models/Item";
import { PriceRequest, PriceRequestId } from "../models/PriceRequest";

export interface PriceRequestRepositoryI {
    postNewRequest(keyword: string): Promise<PriceRequest>
    /// return null if not yet finished
    fetchRequestResults(id: PriceRequestId): Promise<Item[] | null> 
}