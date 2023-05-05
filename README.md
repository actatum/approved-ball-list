# Approved Ball List

[![CircleCI](https://circleci.com/gh/actatum/approved-ball-list/tree/main.svg?style=svg&circle-token=6435d9cf6dd236e074a32c71070bf9d37eafb604)](https://circleci.com/gh/actatum/approved-ball-list/tree/main) [![GoReportCard example](https://goreportcard.com/badge/github.com/actatum/approved-ball-list)](https://goreportcard.com/report/github.com/actatum/approved-ball-list)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/actatum/approved-ball-list.svg)](https://github.com/actatum/approved-ball-list)

Approved Ball List is a discord bot that messages a channel when new balls are added to the [USBC approved ball list](https://www.bowl.com/approvedballlist/)

The bot is deployed as a docker container on Google Cloud Platform's (GCP) Cloud Run. I utilize GCP's Cloud Scheduler to setup a cron schedule to run the bot once every hour. The bot retrieves the list of approved balls from the USBC, filtering for only active brands, and then compares the list to the bot's current database. If there are any balls on the approved ball list from the USBC that aren't in the database they are added and a notification is sent to the discord server.

## Motivation

I'm a fairly active member of Luke Rosdahl's great StormNation discord where despite the name we discuss all things bowling related. From tournaments and bowling balls (of all brands) to the professional tour. It's a great community for bowlers of all levels to get to know and learn from bowlers across the country. As an active member I noticed there were a few people who would fairly regularly check the USBC's approved ball list and update everyone on the new additions. So I decided to write a bot to automate the task and make sure no one misses out on any new additions to the approved ball list!
