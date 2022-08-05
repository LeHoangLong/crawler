import React, { useEffect, useRef, useState } from 'react';
import logo from './logo.svg';
import './App.css';
import { WebSocketI, WebSocketRepository } from './repositories/WebSocketRepository';
import { PriceRequestRepositoryI } from './repositories/PriceRequestRepositoryI'
import { PriceRequestRepositoryWs } from './repositories/PriceRequestRepositoryWs';
import { Loading } from './widgets/Loading';
import { ResultPage } from './widgets/ResultPage';
import { Item } from './models/Item';
import styles from './App.module.scss';
import { SearchBar } from './widgets/SearchBar';

function App() {
  let repository = useRef<WebSocketRepository>(new WebSocketRepository(process.env.REACT_APP_BACKEND_WS_URL!))
  let [priceRequestRepositoryI, setPriceRequestRepositoryI] = useState<PriceRequestRepositoryI | null>(null)
  let [socket, setSocket] = useState<WebSocketI|null>(null)
  let [loading, setIsLoading] = useState<boolean>(false)
  let [prices, setPrices] = useState<Item[]>([])
  let [searchKeyword, setSearchKeyword] = useState<string>("")

  useEffect(() => {
    let socket = repository.current.getNewSocket()
    setSocket(socket)
    return () => {
      socket.close()
    }
  }, [])

  useEffect(() => {
    if (socket !== null) {
      let repo = new PriceRequestRepositoryWs(socket)
      setPriceRequestRepositoryI(repo)
      return () => {
        repo.close()
      }
    }
  }, [socket])


  function searchButtonClicked() {
    if (priceRequestRepositoryI !== null && searchKeyword != "") {
      setIsLoading(true)
      priceRequestRepositoryI.postNewRequest(searchKeyword).then((prices) => {
        setIsLoading(false)
        setPrices(prices.items)
      })
    }
  }

  if (loading) {
    return <div className={ styles.loading }><Loading></Loading></div>
  } else {
    return <div className={ styles.page }>
      <h1 className={ styles.title }>So sánh giá</h1>
      <div className={ styles.search_bar_container }>
        <button className={ styles.search_button } onClick={ searchButtonClicked }>Tìm</button>
        <div className={ styles.search_bar }>
          <SearchBar placeholder="Bạn cần tìm gì" value={ searchKeyword } onChange={ setSearchKeyword }></SearchBar>
        </div> 
      </div>
      <ResultPage results={ prices }></ResultPage>
    </div>
  }

  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.tsx</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header>
    </div>
  );
}

export default App;
