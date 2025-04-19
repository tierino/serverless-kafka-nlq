/// <reference path="./.sst/platform/config.d.ts" />

export default $config({
  app(input) {
    return {
      name: "bikes",
      removal: input?.stage === "production" ? "retain" : "remove",
      protect: ["production"].includes(input?.stage),
      home: "aws",
      providers: { awsx: "2.21.1" },
    }
  },

  async run() {
    const prefixedName = (name: string) => `bikes-${name}`

    const repo = new awsx.ecr.Repository(prefixedName("repo"), {
      forceDelete: true,
    })

    const runQueryImage = new awsx.ecr.Image(prefixedName("runQueryImage"), {
      repositoryUrl: repo.url,
      context: "../../run-query",
    })

    const runQueryRole = new aws.iam.Role(prefixedName("runQueryRole"), {
      assumeRolePolicy: aws.iam.assumeRolePolicyForPrincipal({
        Service: "lambda.amazonaws.com",
      }),
    })

    new aws.iam.RolePolicyAttachment(prefixedName("runQueryLambdaFullAccess"), {
      role: runQueryRole.name,
      policyArn: aws.iam.ManagedPolicy.AWSLambdaExecute,
    })

    new aws.iam.RolePolicyAttachment(prefixedName("s3ReadAccess"), {
      role: runQueryRole.name,
      policyArn: aws.iam.ManagedPolicy.AmazonS3ReadOnlyAccess,
    })

    const bedrockPolicy = new aws.iam.Policy(prefixedName("bedrockAccess"), {
      policy: JSON.stringify({
        Version: "2012-10-17",
        Statement: [
          {
            Action: ["bedrock:InvokeModel"],
            Effect: "Allow",
            Resource:
              "arn:aws:bedrock:*::foundation-model/anthropic.claude-3-5-sonnet-20241022-v2:0",
          },
        ],
      }),
    })

    new aws.iam.RolePolicyAttachment(prefixedName("bedrockAccessAttach"), {
      role: runQueryRole.name,
      policyArn: bedrockPolicy.arn,
    })

    const runQuery = new aws.lambda.Function(prefixedName("runQuery"), {
      packageType: "Image",
      imageUri: runQueryImage.imageUri,
      role: runQueryRole.arn,
      timeout: 30,
      memorySize: 512,
      environment: {
        variables: {
          LAKE_BUCKET_NAME: process.env.LAKE_BUCKET_NAME,
        },
      },
    })

    new aws.lambda.Permission("lambdaPermission", {
      action: "lambda:InvokeFunction",
      principal: "apigateway.amazonaws.com",
      function: runQuery,
    })

    const api = new aws.apigatewayv2.Api(prefixedName("api"), {
      protocolType: "HTTP",
      routeKey: "POST /question",
      target: runQuery.invokeArn,
    })

    return {
      apiEndpoint: api.apiEndpoint,
    }
  },
})
