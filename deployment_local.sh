set -e 

docker build --build-arg ENV_FILE=.env -t deepmarket-local:latest .
docker stop deepmarket-local
docker rm deepmarket-local
docker run -d --name deepmarket-local -p 8080:8080 -it deepmarket-local:latest