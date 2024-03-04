# ChohoCloud Go Sample

Please note: This code is provided solely as a programming reference under GO, and it does not use any third-party libraries. When used in real scenarios, you should call the JSON/HTTP processing methods from the GO framework you use.

**Polling should not be used in the production environment** to determine the completion of tasks as shown in the sample code. Instead, **callbacks should be used** to receive task completion information (i.e., configure the `notification` field in the task startup information).

To get started quickly or view more algorithm invocation examples, we suggest starting with our Python sample to understand HTTP request methods and request parameters: [https://gitee.com/chohotech/api_python_sample](https://gitee.com/chohotech/api_python_sample) (Github: [https://github.com/choho-tech/api_python_sample](https://github.com/choho-tech/api_python_sample))

## Usage Steps

Fill in the corresponding information in the constants section (lines 15-23) in `seg.go` and run in the command line:

```bash
go run seg.go
```

In the same directory as `seg.go`, segmented results `processed_mesh.stl` and `seg_labels.txt` will be generated.

## Example

- This example demonstrates:
  1. How to create a new task JSON
  2. How to create a new task on the server
  3. How to query task status from the server and wait for task completion
  4. How to retrieve task results
  5. How to parse task results
- Please note that while we demonstrate how to perform a segmentation task here, other tasks follow similar patterns, and users can easily adapt them with simple modifications.
- The `main` function in this example demonstrates how to segment an STL jaw file and write the results to disk.

## Code License

This repository is open source under the AGPL v3.0 license. If you use code from this repository in your project, you must provide the source code to users (including SaaS users). If you are a paying customer of Chohotech, this code is licensed to you according to our subscription agreement, and you are not obligated to comply with the AGPL v3.0 open-source license.
