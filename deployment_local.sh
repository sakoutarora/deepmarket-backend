set -e 
docker stop deepmarket-local
docker rm deepmarket-local

docker build --build-arg ENV_FILE=.env -t deepmarket-local:latest .
docker run -d --name deepmarket-local -p 443:443 -it deepmarket-local:latest