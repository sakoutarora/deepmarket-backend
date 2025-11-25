1. Change the backtest structure to support dynamic time series i.e option thing like vwap on atm put + atm close will have to write a custom ohlc data construcutor for this
2. change structure to support multiple set of entry, exit condition
3. change condition to support [position] per [entry,exit condition]
4. posisition to take input from some screener
5. rn the exit is only based on indicator / underlying ts but it can be based on other ts
6. option to add per trade cache and exit based on that
7. option to support multiple time frames concatination