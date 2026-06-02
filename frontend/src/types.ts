export interface Application {
  id: string
  name: string
  repoUrl: string
  branch: string
  composePath: string
  composeContent?: string
  domain: string
  envVars: string
  status: string
  source: string
  createdAt: string
  updatedAt: string
}

export interface Deployment {
  id: string
  applicationId: string
  status: string
  commitSha: string
  commitMessage: string
  logPath: string
  createdAt: string
  updatedAt: string
}

export interface DashboardItem extends Application {
  latestDeployment: Deployment | null
}
