import React from "react"
import { Item } from "../models/Item"
import styles from './ResultPage.module.scss';

export interface ResultPageProps {
    results: Item[]
}

export const ResultPage = (props: ResultPageProps) => {
    let results : React.ReactNode[] = []

    let sortedResults = [...props.results]
    sortedResults.sort((a, b) => {
        let minA = parseInt(a.min.split(" ")[0].split(".").join(""))
        let minB = parseInt(b.min.split(" ")[0].split(".").join(""))
        return minA - minB
    })

    for (let i = 0; i < sortedResults.length; i++) {
        let result = sortedResults[i]
        let matches = result.url.match(/https:\/\/(www.)?([\w\.]*)\/?.*/)!
        let hostname = matches[matches.length - 1]
        results.push(
            <div key={ result.url } className={ styles.result }>
                <a className={ styles.link } target={"_blank"} href={ result.url }>
                    <img src={ result.image }></img>
                    <div className={ styles.content }>
                        <p className={ styles.title }>{ result.title }</p>
                        {(() => {
                            if (result.max == "" ) {
                                return <p className={ styles.price }>{ result.min }</p>
                            } else {
                                return <p className={ styles.price }><span>{ result.min }</span> <span> - </span> <span>{ result.max }</span></p>
                            }
                        })()}
                        <p  className={ styles.hostname }>{ hostname }</p>
                    </div>
                </a>
            </div>
        )
    }

    return <div className={ styles.results }>
        { results }
    </div>
}