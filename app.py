from fastapi import FastAPI, HTTPException, APIRouter
from pydantic import BaseModel
from prophet import Prophet
from holidays import CountryHoliday
import pandas as pd
import logging
import uvicorn
import asyncio

# Initialize logging to show only errors
logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger(__name__)

# Initialize the FastAPI app
app = FastAPI()

# Create a new router for version 1 (v1)
v1_router = APIRouter(prefix="/v1")

# Define the data model for the incoming request using Pydantic
class PredictionRequest(BaseModel):
    expenses: list = None
    incomes: list = None
    savings: dict = None
    start_date: str
    frequency: str = "monthly"  # Default to monthly if not provided
    country: str = "Kenya"  # Default to Kenya if not provided
    prediction_period: int = 3  # Default to 3 periods if not provided
    tax_deductions: bool = False  # Default to no tax deductions if not provided
    tax_rate: float = 0.1  # Default tax rate of 10%
    enable_seasonality: bool = False  # Option to enable seasonality
    enable_holidays: bool = False  # Option to enable holidays

# Function to create a dataframe for Prophet
def create_dataframe(dates, values):
    return pd.DataFrame({'ds': dates, 'y': values})

# Helper to generate country-specific holidays
def get_holidays(country: str):
    try:
        holidays = CountryHoliday(country)
        return pd.DataFrame({
            'ds': list(holidays),
            'holiday': [1] * len(holidays)
        })
    except Exception:
        logger.error(f"Country {country} not supported")
        return pd.DataFrame(columns=['ds', 'holiday'])

# Function to apply tax deductions to incomes
def apply_tax_deductions(incomes: list, tax_rate: float):
    return [income * (1 - tax_rate) for income in incomes]

# General Forecasting Function for Incomes/Expenses with optional seasonality and holidays
def forecast_data(dates, values, country, prediction_period, enable_seasonality, enable_holidays, tax_deductions=False, tax_rate=0.1):
    df = create_dataframe(dates, values)

    # Instantiate a new Prophet model for each request
    model = Prophet()

    # Apply tax deduction if enabled
    if tax_deductions:
        values = apply_tax_deductions(values, tax_rate)

    # Add monthly seasonality if enabled
    if enable_seasonality:
        model.add_seasonality(name='monthly', period=30.5, fourier_order=5)

    # Add holidays if enabled and country is supported
    if enable_holidays:
        holidays_df = get_holidays(country)
        if not holidays_df.empty:
            model = model.add_country_holidays(country_name=country)

    model.fit(df)
    future = model.make_future_dataframe(periods=prediction_period, freq='ME')
    forecast = model.predict(future)
    return forecast[['ds', 'yhat', 'yhat_lower', 'yhat_upper']].tail(prediction_period).to_dict(orient='records')



# Main prediction function using asyncio for parallelism
async def process_predictions(data: PredictionRequest):
    try:
        dates = get_dates(data.start_date, data.frequency, len(data.expenses or []))

        # Run tasks concurrently using asyncio
        tasks = []
        if data.expenses:
            tasks.append(asyncio.to_thread(forecast_data, dates, data.expenses, data.country, data.prediction_period, data.enable_seasonality, data.enable_holidays, data.tax_deductions, data.tax_rate))
        if data.incomes:
            tasks.append(asyncio.to_thread(forecast_data, dates, data.incomes, data.country, data.prediction_period, data.enable_seasonality, data.enable_holidays, data.tax_deductions, data.tax_rate))
        if data.savings:
            tasks.append(asyncio.to_thread(forecast_savings, data.savings, data.start_date, data.prediction_period, data.country, data.enable_seasonality, data.enable_holidays))

        # Wait for all tasks to complete
        results = await asyncio.gather(*tasks)

        # Extract the results
        expense_predictions = results[0] if data.expenses else None
        income_predictions = results[1] if data.incomes else None
        savings_predictions = results[2] if data.savings else None

        return {
            "expense_predictions": expense_predictions,
            "income_predictions": income_predictions,
            "savings_predictions": savings_predictions
        }

    except Exception as e:
        logger.error(f"Prediction processing failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

# Async route for FastAPI to handle incoming predictions
@v1_router.post("/predict")
async def predict_financials(data: PredictionRequest):
    logger.info("Received prediction request")
    try:
        result = await process_predictions(data)
        logger.info("Prediction successfully processed")
        return result
    except Exception as e:
        logger.error(f"Error in prediction endpoint: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

# Include the versioned router in the FastAPI app
app.include_router(v1_router)

# Run the FastAPI app
if __name__ == "__main__":
    uvicorn.run("app:app", host="0.0.0.0", port=8000, reload=True)
