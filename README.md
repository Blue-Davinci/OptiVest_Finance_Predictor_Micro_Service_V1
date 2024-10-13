# Financial Prediction API

## Overview

This is a FastAPI microservice designed to work with the Optivest AI financial advisor. 
It provides endpoints for predicting financial metrics like expenses, incomes, and savings using time series forecasting with Prophet.

## Endpoints

### POST `/v1/predict`

Predict financial metrics based on the provided data.

#### Request Body

```json
{
    "expenses": [
        1000,
        1200,
        1100,
        1050,
        1300
    ],
    "expenses_start_date": "2024-03-01",
    "incomes": [
        3000,
        3200,
        3100,
        3000,
        3300
    ],
    "incomes_start_date": "2024-04-01",
    "savings": {
        "current_savings": 10000,
        "monthly_contribution": 3000,
        "goal": 20000
    },
    "savings_start_date": "2024-05-01",
    "frequency": "monthly",
    "country": "Kenya",
    "prediction_period": 3,
    "enable_tax_deductions": false,
    "tax_rate": 0.1,
    "enable_seasonality": false,
    "enable_holidays": false
}

```
### Values
- expenses: List of historical expense values.
- incomes: List of historical income values.
- savings: Object containing current savings, monthly contribution, and goal.
- start_date: Starting date for predictions.
- frequency: Frequency of predictions (e.g., "monthly").
- country: Country for which holidays are considered.
- prediction_period: Number of future periods to predict.
- tax_deductions: Boolean to indicate if tax deductions should be applied.
- tax_rate: Tax rate to apply if tax deductions are enabled.
- enable_seasonality: Boolean to enable seasonality in the prediction model.
- enable_holidays: Boolean to include holidays in the prediction model.

### Response
```json
{
  "expense_predictions": [
    {"ds": "2023-04-30", "yhat": 180, "yhat_lower": 150, "yhat_upper": 210},
    {"ds": "2023-05-31", "yhat": 190, "yhat_lower": 160, "yhat_upper": 220},
    {"ds": "2023-06-30", "yhat": 200, "yhat_lower": 170, "yhat_upper": 230}
  ],
  "income_predictions": [
    {"ds": "2023-04-30", "yhat": 230, "yhat_lower": 200, "yhat_upper": 260},
    {"ds": "2023-05-31", "yhat": 240, "yhat_lower": 210, "yhat_upper": 270},
    {"ds": "2023-06-30", "yhat": 250, "yhat_lower": 220, "yhat_upper": 280}
  ],
  "savings_predictions": [
    {
      "ds": "2023-04-30",
      "yhat": 1200,
      "yhat_lower": 1100,
      "yhat_upper": 1300,
      "goal_met": "No",
      "surplus_or_deficit": -3800
    },
    {
      "ds": "2023-05-31",
      "yhat": 1400,
      "yhat_lower": 1300,
      "yhat_upper": 1500,
      "goal_met": "No",
      "surplus_or_deficit": -3600
    },
    {
      "ds": "2023-06-30",
      "yhat": 1600,
      "yhat_lower": 1500,
      "yhat_upper": 1700,
      "goal_met": "No",
      "surplus_or_deficit": -3400
    }
  ]
}
```

### Prerequisites

- Python 3.10
- `pip` (Python package installer)

### Installation
1. **Clone the repository:**

   ```sh
   git clone https://github.com/Blue-Davinci/OptiVest_Finance_Predictor_Micro_Service_V1
   cd OptiVest_Finance_Predictor_V1
  ```
To run the micro-service API, use the following command:

2. **Create a virtual environment with Python 3.10:**
```sh
C:\Users\KLAUS\AppData\Local\Programs\Python\Python310\python.exe -m venv .venv
```

3. **Activate the virtual environment:**
```bash
& ".\.venv\Scripts\Activate.ps1"
```

4. **Install the dependencies:**
```bash
pip install -r requirements.txt
```

5. **Run the FastAPI application:**
```bash
uvicorn app:app --host 0.0.0.0 --port 8000 --reload
```

### Dependancies
- FastAPI
- Uvicorn
- Pydantic
- Prophet
- pandas
- holidays
- asyncio
- logging

**License**
This project is licensed under the MIT License. See the LICENSE file for details.

And awaaaay you go! 🚀
