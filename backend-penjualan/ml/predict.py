import sys
import json
import os
import pandas as pd
import numpy as np
from prophet import Prophet
from prophet.serialize import model_from_json, model_to_json
import warnings

# Suppress warning plotly yang tidak penting
warnings.filterwarnings("ignore", message="Importing plotly failed")

MODEL_PATH = 'models/prophet_model.json'

# Buat model dasar + tambahkan holidays & regressor DULU
model = Prophet(
    weekly_seasonality=True,
    yearly_seasonality=True,
    changepoint_prior_scale=0.8,
    seasonality_prior_scale=10,
    uncertainty_samples=200
)
model.add_country_holidays(country_name='ID')
model.add_regressor('is_weekend')

retrain = False

# Coba load model yang sudah ada
try:
    if os.path.exists(MODEL_PATH):
        with open(MODEL_PATH, 'r') as f:
            model = model_from_json(f.read())
except Exception:
    retrain = True

# Input dari Go
try:
    input_data = json.load(sys.stdin)
    csv_path = input_data.get('csv_path')
    periods = input_data.get('periods', 30)
except:
    print(json.dumps({'error': 'Invalid input'}), flush=True)
    sys.exit(1)

if not csv_path or not os.path.exists(csv_path):
    print(json.dumps({'error': 'CSV not found'}), flush=True)
    sys.exit(1)

# Load CSV
df = pd.read_csv(csv_path)

# Deteksi kolom date dan value
date_col = next((c for c in df.columns if c.lower() in ['date', 'tanggal', 'ds']), df.columns[0])
value_col = next((c for c in df.columns if c.lower() in ['projected_quantity', 'value', 'penjualan', 'y', 'quantity']), df.columns[1])

df['ds'] = pd.to_datetime(df[date_col])
df['y'] = pd.to_numeric(df[value_col], errors='coerce')
df = df[['ds', 'y']].dropna()

if len(df) < 2:
    print(json.dumps({'error': 'Data too small (<2 rows)'}), flush=True)
    sys.exit(1)

df = df.sort_values('ds').reset_index(drop=True)

# Tambah regressor is_weekend
df['is_weekend'] = (df['ds'].dt.dayofweek >= 5).astype(int)
prophet_df = df[['ds', 'y', 'is_weekend']].copy()

# Jika gagal load atau model belum ada → retrain
if retrain:
    model.fit(prophet_df)
    os.makedirs('models', exist_ok=True)
    with open(MODEL_PATH, 'w') as f:
        f.write(model_to_json(model))

# Predict
future = model.make_future_dataframe(periods=periods)
future['is_weekend'] = (future['ds'].dt.dayofweek >= 5).astype(int)
forecast = model.predict(future)

hist_len = len(prophet_df)

# Historical: data masa lalu dengan y aktual + fitted values
historical = pd.DataFrame({
    'ds': forecast['ds'][:hist_len].dt.strftime('%Y-%m-%d').tolist(),
    'y': prophet_df['y'].round(0).astype(int).tolist(),  # nilai aktual
    'yhat': np.round(forecast['yhat'][:hist_len]).astype(int).tolist(),
    'yhat_lower': np.floor(forecast['yhat_lower'][:hist_len]).astype(int).tolist(),
    'yhat_upper': np.ceil(forecast['yhat_upper'][:hist_len]).astype(int).tolist()
}).to_dict('records')

# Forecast: hanya masa depan
forecast_data = pd.DataFrame({
    'ds': forecast['ds'][hist_len:].dt.strftime('%Y-%m-%d').tolist(),
    'yhat': np.round(forecast['yhat'][hist_len:]).astype(int).tolist(),
    'yhat_lower': np.floor(forecast['yhat_lower'][hist_len:]).astype(int).tolist(),
    'yhat_upper': np.ceil(forecast['yhat_upper'][hist_len:]).astype(int).tolist()
}).to_dict('records')

# Output JSON yang benar
result = {
    "historical": historical,
    "forecast": forecast_data
}

print(json.dumps(result), flush=True)