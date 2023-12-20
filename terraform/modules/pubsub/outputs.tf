output "topic" {
  value = module.pubsub.topic
}

output "dlq" {
  value = module.dlq.topic
}
