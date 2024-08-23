Docker Volume Driver for Amazon S3 using Mountpoint S3
=============================================

The purpose of this project is to provide a Docker Volume Driver for Mounting Amazon S3 Buckets using Mountpoint S3.  

The idea it to provide a simple way to mount S3 buckets to Docker containers. 

The plugin image is built around Alpine Linux. 

## Installation
To install the plugin you need to run the following command:

    docker plugin install aekis/docker-mount-s3

Plugin "aekis/docker-mount-s3" is requesting the following privileges:

    - network: [host]
    - device: [/dev/fuse]
    - capabilities: [CAP_SYS_ADMIN]
    Do you grant the above permissions? [y/N]

Accept the permissions by typing `y` and pressing `Enter`.

You could also set an alias for the plugin by using the following command:

    docker plugin install --alias mount-s3 aekis/docker-mount-s3 

and then you could use the alias as the driver to create the volumes. 

## Docker Compose
The solution provided here is to use a single plugin and use the driver and driver_opts to provide the credentials and options to mount-s3 setting up the environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY so mount-s3 can use them to mount the bucket.

Example usage in a compose.yaml file:

    volumes:
      volume_name:
        driver: aekis/docker-mount-s3
        driver_opts:
          bucket: bucket_name
          AWS_ACCESS_KEY_ID: ${AWS_ACCESS_KEY_ID}
          AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY}
          o: --allow-delete --allow-overwrite --allow-other --region=us-east-1 
For the values of the environment variables you could use the .env file to set them and use them in the compose.yaml file.

Example of .env file:

    AWS_ACCESS_KEY_ID=AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY=AWS_SECRET_ACCESS_KEY

That .env file should be in the same directory as the compose.yaml file and the `docker compose up` command will use it to set the environment variables.

You could directly set the credentials in the compose.yaml file.

    volumes:
      volume_name:
        driver: aekis/docker-mount-s3
        driver_opts:
          bucket: bucket_name
          AWS_ACCESS_KEY_ID: XXXXXXAWS_ACCESS_KEY_IDXXXXXXX
          AWS_SECRET_ACCESS_KEY: XXXXXXXXXXXXAWS_SECRET_ACCESS_KEYXXXXXXXXXXXX
          o: --allow-delete --allow-overwrite --allow-other --region=us-east-1


## Docker Volume Create
You could use the following command to manually create the volume:
    
    docker volume create --driver aekis/docker-mount-s3 --opt bucket=bucket --opt AWS_ACCESS_KEY_ID=AWS_ACCESS_KEY_ID --opt AWS_SECRET_ACCESS_KEY=AWS_SECRET_ACCESS_KEY --opt o="--allow-delete --allow-overwrite --allow-other --region=us-east-1" volume_name

## Motivation

Existing Solutions involves creating one plugin per AWS Credentials, which is not practical when you need to manage multiple AWS Accounts.

With those solutions you can't use the same plugin for all of them and you need to create a new plugin for each account and match the pluging alias with the credentials that you wanna use. 

Those plugins are not flexible enough to allow you to use environment variables to provide the credentials when creating the volume for example in docker compose. 

They defends that the credentials should be provided in the plugin configuration by setting the environment variables on the plugin but that will save the credentials in the plugin configuration which is not secure neither because you could inspect the plugin to see the credentials.