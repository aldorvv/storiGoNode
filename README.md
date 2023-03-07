# Stori challenge

Same API written in Golang!

Using the same stack as the Python example (S3 Bucket, MySQL and Docker) this
API provides you a simple way to upload a CSV file to a local S3 and get a tidy
summary in your inbox!

## Setup

First of all, clone this repo!
´´´bash
git clone git@github.com:aldorvv/storiGoNode.git
cd storiGoNode
´´´

In order to setup the email account to use you should use an "Application password"
for your gmail account.

You can see [how to get one here!](https://support.google.com/accounts/answer/185833?)

Once you got your password and user, create a copy of ´.env.template´ file with
´´´bash
cp .env.template .env
´´´

and add the values of your username and password at ´EMAIL_HOST_USER´ and ´EMAIL_HOST_PASSWORD´
respectively.

## Up and run!

Once you did all the setup steps, please go ahead and up the server with
´´´bash
docker compose up
´´´
Please be patient! It'll take about a couple of minutes to start, then please go ahead and test it as you please,
I already attached a sample csv file named ´out.csv´ and a postman collection for make your testing easier.
