from abc import ABC, abstractmethod
import pandas as pd
import numpy as np
from binance.client import Client
from enums import Interval

class DataIngestor(ABC):
    @abstractmethod
    def get_data(self, interval: str, symbol: str, start_date: str, end_date: str):
        pass

class BinanceDataIngestor(DataIngestor):
    def __init__(self, api_key : str, api_secret : str):
        self.api_key = api_key
        self.api_secret = api_secret
        self.client = Client(self.api_key, self.api_secret)        
        self.df = pd.DataFrame()

    def get_data(self, interval: Interval, symbol: str, start_date: str, end_date: str):
        columns = [
        "open_time", "open", "high", "low", "close", "volume",
        "close_time", "quote_asset_volume", "number_of_trades",
        "taker_buy_base_asset_volume", "taker_buy_quote_asset_volume", "ignore"
    ]    
        df = pd.DataFrame(self.client.get_historical_klines(symbol, interval.value, start_date, end_date), columns=columns)

        return df

        