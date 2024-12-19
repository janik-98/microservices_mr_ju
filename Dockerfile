FROM ubuntu:22.04

RUN apt-get update && apt-get install -y python3 python3-pip

RUN pip install flask

COPY app.py /app/app.py

CMD FLASK_APP=/app/app.py flask run --host=0.0.0.0 --port=8080