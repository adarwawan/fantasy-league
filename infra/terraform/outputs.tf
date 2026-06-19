output "database_url" {
  value     = neon_project.main.connection_uri
  sensitive = true
}

output "redis_url" {
  value     = "rediss://:${upstash_redis_database.cache.password}@${upstash_redis_database.cache.endpoint}:${upstash_redis_database.cache.port}"
  sensitive = true
}

output "api_ip" {
  value = var.fly_api_ip != "" ? var.fly_api_ip : "(not yet set — run fly ips allocate-v4 --app fantasy-league-api)"
}
