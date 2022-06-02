docker network create -d bridge roachnet
docker run -d --name=roach1 --hostname=roach1 --net=roachnet -p 26257:26257 -p 8080:8080  -v "${PWD}/cockroach-data/roach1:/cockroach/cockroach-data"  cockroachdb/cockroach:v22.1.0 start --insecure --join=roach1,roach2,roach3
sleep 1
docker run -d --name=roach2 --hostname=roach2 --net=roachnet -v "${PWD}/cockroach-data/roach2:/cockroach/cockroach-data" cockroachdb/cockroach:v22.1.0 start --insecure --join=roach1,roach2,roach3
sleep 1
docker run -d --name=roach3 --hostname=roach3 --net=roachnet -v "${PWD}/cockroach-data/roach3:/cockroach/cockroach-data" cockroachdb/cockroach:v22.1.0 start --insecure --join=roach1,roach2,roach3
sleep 5s
docker exec -it roach1 ./cockroach init --insecure
# docker exec -it roach1 ./cockroach sql --insecure