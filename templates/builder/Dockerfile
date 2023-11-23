# The following Dockerfile is simply an EXAMPLE of what a builder dockerfile may look like.
# It should not be used in production anywhere, by anyone. It exists for testing purposes only.

FROM monarch:latest

WORKDIR /app

COPY royal.yaml .
COPY example-service.py .
COPY requirements.txt .

# must specify full paths
CMD [ "/app/royal.yaml", "/app/example-service.py", "/app/requirements.txt" ]