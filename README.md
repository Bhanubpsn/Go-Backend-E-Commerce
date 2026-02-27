Hey if you are reading this then here are the specifics to run this project.

If you want to run the DB on a container => Install Docker Desktop.
Install Go.

1. Spin up the Docker containers to run the MongoDB instances, there should be 3 instances created 1 primary, 2 secondary.
   docker compose up -d

2. Run Web Servers (at max right now 3 allowed) on different PORTS (choose them according to your need just update the env).
   Navigate inside Backend folder and run
   go mod tidy (for the first time)
   go run main.go

3. Run the Load Balancer on the PORT of your choice just don't let the PORTS clash with each other.
   Navigate to LoadBalancer folder and run
   go mod tidy (for the first time)
   go run main.go

4. Do the same for Rate Limiter, run on a different port.
   Navigate to RateLimiter and run
   go mod tidy (for the first time)
   go run main.go

5. For Message broker first add your email and password in the env of those folders, not your actual password of your email, Go to Google Accounts => Add Passwords => A 16 character long random password will be generated for you, use that.
   Why all this?
   Because here I have used SMTP for email service which is an old service so Google Auth can't be cleared by some old services thats why to run those services Google uses this App passwords to by pass the AUTH (2-Step-Verification) and directly use your account. Its very unsecure to be honest but just for the sake of this LOCAL project we can use that.

6. Inside the MessageBroker folder first run the Broker then the Worker.
   go mod tidy (for the first time)
   go run main.go
   In both the folders.

7. curl your requests or use Postman. Hit the Loadbalancer PORT not the actual server PORT.

8. Look out for the logs.

Thank you! ^\_^
