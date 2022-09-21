resource "solidserver_device" "my_first_device" {
  name   = "my_device"
  class  = "AWS_EC2_INSTANCE"
  class_parameters = {
    cloudaz = "eu-west-1a"
    instanceid = "i-03d4bd36f915b0322"
    instancetype = "t2.micro"
  }
}