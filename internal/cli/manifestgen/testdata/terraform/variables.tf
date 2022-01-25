variable "name" {
  type        = string
  description = "Name of environment"
  default     = "test"
}

variable "count" {
  type        = number
  description = "Number of instances"
  default     = 3
}
