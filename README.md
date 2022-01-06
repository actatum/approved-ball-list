# Approved Ball List
[![actatum](https://circleci.com/gh/actatum/approved-ball-list.svg?style=svg)](https://app.circleci.com/pipelines/github/actatum/approved-ball-list?filter=all) [![GoReportCard example](https://goreportcard.com/badge/github.com/actatum/approved-ball-list)](https://goreportcard.com/report/github.com/actatum/approved-ball-list)

Approved Ball List is a discord bot that messages a channel when new balls are added to the [USBC approved ball list](https://www.bowl.com/approvedballlist/)

The approved ball list runs as two parts deployed as separate cloud functions on Google Cloud Platform. The first cloud functions is triggered on a cron schedule every day at Noon eastern time. This function retrieves the entire list of approved balls from the USBC, filters out inactive brands and then compares the list to the current database. If there are any balls on the approved ball list that aren't in the database they are added. The second cloud function is triggered from firestore write events and for each new ball it will send a discord message to the configured channels.

## Motivation

I'm a fairly active member of Luke Rosdahl's great StormNation discord where despite the name we discuss all things bowling related. From tournaments and bowling balls (of all brands) to the professional tour. It's a great community for bowlers of all levels to get to know and learn from bowlers across the country. As an active member I noticed there were a few people who would fairly regularly check the USBC's approved ball list and update everyone on the new additions. So I decided to write a bot to automate the task and make sure no one misses out on any new additions to the approved ball list!