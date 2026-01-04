FROM python:3.9-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    zlib1g-dev \
    libjpeg-dev \
    && rm -rf /var/lib/apt/lists/*

# Clone repository
RUN git clone https://github.com/petrov-e/neo4j_exporter.git /app

WORKDIR /app

# Pin Werkzeug to avoid ImportError with Flask 2.1.2
RUN echo "Werkzeug==2.2.2" >> src/requirements.txt

# Install python dependencies
RUN pip install --no-cache-dir -r src/requirements.txt

# Set PYTHONPATH
ENV PYTHONPATH=/app/src

# Run application
# Patch app.py to listen on 0.0.0.0:9347 instead of default 127.0.0.1:5000
RUN sed -i "s/app.run(debug=True)/app.run(host='0.0.0.0', port=9347)/" src/app.py

# Run application
CMD ["python", "src/app.py"]
