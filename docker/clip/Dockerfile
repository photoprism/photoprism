# copied from https://fastapi.tiangolo.com/deployment/docker/#docker-image-with-poetry

# get requirements.txt from poetry
FROM python:3.10 as requirements-stage
WORKDIR /tmp
RUN pip install poetry
COPY pyproject.toml poetry.lock* /tmp/
RUN poetry export -f requirements.txt --output requirements.txt --without-hashes

# install requirements.txt
FROM python:3.10
WORKDIR /code
COPY --from=requirements-stage /tmp/requirements.txt /code/requirements.txt
RUN pip install --no-cache-dir --upgrade -r /code/requirements.txt

# copy app
COPY clip_api /code/app

# run one encode_text, to force the download of the clip model
RUN python app/clip_api.py 

# run the server on port 8000
CMD ["uvicorn", "app.clip_api:app", "--host", "0.0.0.0"]
