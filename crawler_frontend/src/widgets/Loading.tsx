import { useEffect, useState } from "react"
import { ClipLoader } from "react-spinners"
import styles from "./Loading.module.scss"

export const Loading = () => {
    let [time, setTime] = useState(0)
    useEffect(() => {
        let interval = setInterval(() => {
            setTime((val) => val + 1)
        }, 1000)

        return () => {
            clearInterval(interval)
        }
    }, [])

    return <div className={ styles.loading }>
        <ClipLoader color="#0352fc" loading={true} size={150} />
        <p className={ styles.time }>{time}</p>
    </div>
}