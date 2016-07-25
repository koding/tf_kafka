resource "aws_elastic_beanstalk_environment" "environment" {
  name                   = "${var.env_name}"
  application            = "${var.app_name}"
  solution_stack_name    = "${var.solution_stack_name}"
  wait_for_ready_timeout = "20m"
  tier = "WebServer"

  # ASG
  setting {
    namespace = "aws:autoscaling:asg"
    name      = "Cooldown"
    value     = "360"
  }
  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MinSize"
    value     = "3"
  }
  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MaxSize"
    value     = "4"
  }

  # Launch Configuration

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "EC2KeyName"
    value     = "${var.key_pair_name}"
  }
  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = "${var.iaminstanceprofile}"
  }
  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "ImageId"
    value     = "${var.ami_id}"
  }
  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "${var.instance_type}"
  }
  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "MonitoringInterval"
    value     = "5 minutes"
  }
  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = "${aws_security_group.eb.id}"
  }

  # AWS:AUTOSCALING:TRIGGER

  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "BreachDuration"
    value     = "5"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "LowerBreachScaleIncrement"
    value     = "-1"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "LowerThreshold"
    value     = "6"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "MeasureName"
    value     = "CPUUtilization"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "Period"
    value     = "1"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "Statistic"
    value     = "Average"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "Unit"
    value     = "Percent"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "UpperBreachScaleIncrement"
    value     = "1"
  }
  setting {
    namespace = "aws:autoscaling:trigger"
    name      = "UpperThreshold"
    value     = "40"
  }

  # AWS:AUTOSCALING:UPDATEPOLICY:ROLLINGUPDATE
  # setting {
  #   namespace = "aws:autoscaling:updatepolicy:rollingupdate"
  #   name      = "MaxBatchSize"
  #   value     = "One-third of the minimum size of the autoscaling group, rounded to the next highest integer."
  # }
  # setting {
  #   namespace = "aws:autoscaling:updatepolicy:rollingupdate"
  #   name      = "MinInstancesInService"
  #   value     = "The minimum size of the AutoScaling group or one less than the maximum size of the autoscaling group, whichever is lower."
  # }
  setting {
    namespace = "aws:autoscaling:updatepolicy:rollingupdate"
    name      = "RollingUpdateEnabled"
    value     = "true"
  }
  setting {
    namespace = "aws:autoscaling:updatepolicy:rollingupdate"
    name      = "RollingUpdateType"
    value     = "Health"
  }
  # setting {
  #   namespace = "aws:autoscaling:updatepolicy:rollingupdate"
  #   name      = "PauseTime"
  #   value     = "Automatically computed based on instance type and container."
  # }
  setting {
    namespace = "aws:autoscaling:updatepolicy:rollingupdate"
    name      = "Timeout"
    value     = "PT30M"
  }

  # AWS:EC2:VPC
  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = "${var.vpc_id}"
  }
  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = "${var.vpc_subnets}"
  }
  setting {
    namespace = "aws:ec2:vpc"
    name      = "ELBSubnets"
    value     = "${var.elb_subnet_ids}"
  }
  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  # AWS:ELASTICBEANSTALK:APPLICATION
  setting {
    namespace = "aws:elasticbeanstalk:application"
    name      = "Application Healthcheck URL"
    value     = "${var.healthCheckEndpoint}"
  }

  # AWS:ELASTICBEANSTALK:APPLICATION:ENVIRONMENT
  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "EB_ENV_NAME"
    value     = "koding-kafka"
  }

  # AWS:ELASTICBEANSTALK:COMMAND
  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "DeploymentPolicy"
    value     = "${var.deployment_type}"
  }
  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "Timeout"
    value     = "${var.deployment_command_timeout}"
  }
  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "BatchSizeType"
    value     = "${var.deployment_batch_size_type}"
  }
  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "BatchSize"
    value     = "${var.deployment_batch_size}"
  }
  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "IgnoreHealthCheck"
    value     = "false"
  }
  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "HealthCheckSuccessThreshold"
    value     = "Ok"
  }
  # AWS:ELASTICBEANSTALK:ENVIRONMENT
  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "EnvironmentType"
    value     = "LoadBalanced"
  }
  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = "${var.service_role_name}"
  }

  # AWS:ELASTICBEANSTALK:HEALTHREPORTING:SYSTEM
  setting {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name      = "SystemType"
    value     = "basic"
  }

  # AWS:ELASTICBEANSTALK:SNS:TOPICS
  setting {
    namespace = "aws:elasticbeanstalk:sns:topics"
    name      = "Notification Endpoint"
    value     = "sysops+kafka@koding.com"
  }
  setting {
    namespace = "aws:elasticbeanstalk:sns:topics"
    name      = "Notification Protocol"
    value     = "email"
  }

  # AWS:ELB:HEALTHCHECK
  setting {
    namespace = "aws:elb:healthcheck"
    name      = "HealthyThreshold"
    value     = "3"
  }
  setting {
    namespace = "aws:elb:healthcheck"
    name      = "Interval"
    value     = "10"
  }
  setting {
    # HealthCheck timeout must be less than interval.
    namespace = "aws:elb:healthcheck"
    name      = "Timeout"
    value     = "5"
  }
  setting {
    namespace = "aws:elb:healthcheck"
    name      = "UnhealthyThreshold"
    value     = "5"
  }

  # AWS:ELB:LOADBALANCER
  setting {
    namespace = "aws:elb:loadbalancer"
    name      = "CrossZone"
    value     = "true"
  }
  setting {
    namespace = "aws:elb:loadbalancer"
    name      = "SecurityGroups"
    value     = "${aws_security_group.elb.id}"
  }
  setting {
    namespace = "aws:elb:loadbalancer"
    name      = "ManagedSecurityGroup"
    value     = "${aws_security_group.elb.id}"
  }

  # AWS:ELB:LISTENER
  ## 9092
  setting {
    namespace = "aws:elb:listener:9092"
    name      = "ListenerProtocol"
    value     = "TCP"
  }
  setting {
    namespace = "aws:elb:listener:9092"
    name      = "InstancePort"
    value     = "9092"
  }
  setting {
    namespace = "aws:elb:listener:9092"
    name      = "InstanceProtocol"
    value     = "TCP"
  }
  setting {
    namespace = "aws:elb:listener:9092"
    name      = "ListenerEnabled"
    value     = "true"
  }

  # AWS:ELB:POLICIES
  setting {
    namespace = "aws:elb:policies"
    name      = "ConnectionDrainingEnabled"
    value     = "true"
  }
  setting {
    namespace = "aws:elb:policies"
    name      = "ConnectionDrainingTimeout"
    value     = "20"
  }
  # setting {
  #   namespace = "aws:elb:policies"
  #   name      = "LoadBalancerPorts"
  #   value     = ":all"
  # }
  setting {
    namespace = "aws:elb:policies"
    name      = "ConnectionSettingIdleTimeout"
    value     = "60"
  }
  # setting {
  #   namespace = "aws:elb:policies"
  #   name      = "Stickiness Cookie Expiration"
  #   value     = "0"
  # }
  # setting {
  #   namespace = "aws:elb:policies"
  #   name      = "Stickiness Policy"
  #   value     = "false"
  # }

  # AWS:ELB:POLICIES
  # setting {
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name      = "LoadBalancerPorts"
  #   value     = "60"
  # }
  # setting {
  #   # The name of the application-generated cookie that controls the session
  #   # lifetimes of a AppCookieStickinessPolicyType policy. This policy can be
  #   # associated only with HTTP/HTTPS listeners.
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name  = "CookieName"
  #   value = "None"
  # }
  # setting {
  #   # A comma-separated list of the instance ports that this policy applies to.
  #   # use :all to indicate all instance ports
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name  = "InstancePorts"
  #   value = "79"
  # }
  # setting {
  #   # A comma-separated list of the listener ports that this policy applies to.
  #   # use :all to indicate all instance ports
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name  = "LoadBalancerPorts"
  #   value = "None"
  # }
  # setting {
  #   # For a ProxyProtocolPolicyType policy, specifies whether to include the IP
  #   # address and port of the originating request for TCP messages. This policy
  #   # can be associated only with TCP/SSL listeners.
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name      = "ProxyProtocol"
  #   value     = "true"
  # }
  # setting {
  #   # A comma-separated list of policy names (from the PublicKeyPolicyType
  #   # policies) for a BackendServerAuthenticationPolicyType policy that controls
  #   # authentication to a back-end server or servers. This policy can be
  #   # associated only with back-end servers that are using HTTPS/SSL.
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name      = "PublicKeyPolicyNames"
  #   value     = "None"
  # }
  # setting {
  #   # A comma-separated list of SSL protocols to be enabled for a
  #   # SSLNegotiationPolicyType policy that defines the ciphers and protocols
  #   # that will be accepted by the load balancer. This policy can be associated
  #   # only with HTTPS/SSL listeners.
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name      = "SSLProtocols"
  #   value     = "None"
  # }
  # setting {
  #   # The name of a predefined security policy that adheres to AWS security best
  #   # practices and that you want to enable for a SSLNegotiationPolicyType
  #   # policy that defines the ciphers and protocols that will be accepted by the
  #   # load balancer. This policy can be associated only with HTTPS/SSL
  #   # listeners.
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name      = "SSLReferencePolicy"
  #   value     = "None"
  # }
  # setting {
  #   # The amount of time, in seconds, that each cookie is valid. 0 to 1000000
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name      = "Stickiness Cookie Expiration"
  #   value     = "0"
  # }
  # setting {
  #   # Binds a user's session to a specific server instance so that all requests
  #   # coming from the user during the session are sent to the same server
  #   # instance.
  #   namespace = "aws:elb:policies:ProxyProtocolPolicyType"
  #   name      = "Stickiness Policy"
  #   value     = "false"
  # }

  tags {
    Name = "${var.env_name}"
    monitoring = "datadog"
  }
}
