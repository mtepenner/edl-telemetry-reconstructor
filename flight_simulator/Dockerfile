FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app/ ./app/

# Run the simulator
CMD ["python", "-m", "app.main", "--kafka", "--kafka-broker", "kafka:9092", "--verbose"]
