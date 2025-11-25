DeepMarket Backend

🚀 Overview
The DeepMarket Backend is a powerful system designed to process and execute complex, user-defined algorithmic trading strategies. It takes a declarative strategy definition in JSON format, transforms it into an optimized execution flow (a Directed Acyclic Graph or DAG), and runs it against historical or real-time time-series data fetched from a data source (DBC).

This project acts as the core engine for backtesting and potentially live execution of trading logic, allowing users to define intricate entry and exit conditions using a combination of technical indicators, functions, and logical operators.

✨ Key Features
Declarative Strategy Definition: Strategies are defined using a structured, easy-to-read JSON format.

Abstract Syntax Tree (AST) Generation: The JSON input is parsed into a robust AST for syntax validation and structured representation of the conditions.

Optimized Execution DAG: The AST is transformed into a Directed Acyclic Graph (DAG) to ensure efficient and parallelized calculation of indicators and conditions.

Flexible Indicator Engine: Supports a wide range of indicators (EMA, RSI, SMA, etc.) with customizable parameters and timeframes.

Database Integration: Designed to interface with a dedicated DataBase Client (DBC) for high-speed retrieval of time-series data.

🛠️ Core Workflow
The backend executes a strategy through the following steps:

JSON Input: Receives the strategy definition via an API payload.

AST Creation: Parses the entry_conditions and exit_conditions tokens into a hierarchical Abstract Syntax Tree.

DAG Generation: Optimizes the AST into a DAG, identifying dependencies between indicator calculations (e.g., ensuring an EMA is calculated before it's used in a comparison).

Data Fetching (DBC): Uses the strategy's defined symbol and timeframes to query the necessary time-series data from the DBC.

Execution Engine: Traverses the DAG, calculates the required indicators on the fetched data, and evaluates the entry/exit conditions.
