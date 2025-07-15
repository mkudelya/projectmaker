# ProjectMaker

## Description

The program can be used when there is a large project (for example, microservices) that are located in different directories.
Commands are run in all specified projects

## Use examples:

### Create new_task_branch from main branch
```Bash
projectmaker newtask project1,project2 new_task_branch 
```

### Create new tag in main branch

```Bash
projectmaker newtag project1,project2 
```

### Create new merge request from source_project_branch

```Bash
projectmaker newmergerequest project1,project2 source_project_branch "Task title" 
```