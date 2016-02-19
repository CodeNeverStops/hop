# hop
A go program is used to post back tracking data to HasOffers.

### Usage
./hoff_postback -h
Usage of ./hoff_postback:
-admin-port=8888: admin port                // Manages server's port, default port is: 8888
-fetch-timeout=10: timeout to fetch a task  // Get number of seconds for timeout with a task. Default timeout is 10 seconds
-i=10: retry interval, unit: Second         // Retry interval measured in Seconds. Default is 10 seconds
-log-buffer=2: log buffer size              // Buffer size for log output. Default setting is output logs when there are 2 logs in the buffer
-log-queue=1000: log queue size             // Log queue size, default is 1000 logs
-n=1000: max pool size of workers           // Maximum number of workers, default is 1000
-redis-db=0: redis db                       // Redis DB number, default is 0
-redis-host="127.0.0.1": redis host         // Redis service's host, default is 127.0.0.1
-redis-key="": redis key                    // Redis key.The key's structure is Redis List
-redis-port=6379: redis port                // Redis service's port number, default is 6379
-redis-pwd="": redis password               // Redis service's password
-t=5: retry times                           // Number of times to retry a task, default is 5 times
-v=true: log level 1                        // Log level 1. If enabled, will output Error and Warning related logs. Default level is log level 1
-vv=false: log level 2                      // Log level 2. If enabled, will output Notice and Info related logs, default is off
-vvv=false: log level 3                     // Log level 3. If enabled, will output Debug related logs, default is off.

### Run
./hoff_postback -fetch-timeout=30 -i=300 -n=1000 -redis-db=0 -redis-host=127.0.0.1 -redis-key=hoff:postback:go
