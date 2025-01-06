package handler

import (
	"github.com/gin-gonic/gin"
	"goDump/config"
	k "goDump/kubernetes"
	"io/ioutil"
	"net/http"
	"strings"
)

type podContainerList struct {
	Namespace  string   `json:"namespace"`
	PodName    string   `json:"podname"`
	Containers []string `json:"containers"`
}

func ViewGet(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, embeddedHTML(strings.Join(config.Namespace, ",")))
}

func ViewPost(c *gin.Context) {
	var pod config.CheckPod
	if err := c.ShouldBindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.CheckPodList = append(config.CheckPodList, pod)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func ViewFiles(c *gin.Context) {
	files, err := ioutil.ReadDir(config.ShareDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, config.DownloadHost+"/download/"+file.Name())
		}
	}
	c.JSON(http.StatusOK, fileNames)
}

func ItemInfo(c *gin.Context) {
	ns := c.Query("namespace")
	var pcList []podContainerList
	if ns != "" {
		kubeClient, err := k.NewKubeClient(k.RestConfig())
		if err != nil {
			panic(err)
		}
		defer kubeClient.Close()
		podList, err := k.GetPodList(kubeClient.Client, ns)
		if err != nil {
			panic(err)
		}
		for _, pod := range podList.Items {
			var containerNames []string
			for _, container := range pod.Spec.Containers {
				containerNames = append(containerNames, container.Name)
			}
			pcList = append(pcList, podContainerList{Namespace: pod.Namespace, PodName: pod.Name, Containers: containerNames})
		}
	}
	c.JSON(http.StatusOK, gin.H{"viewPod": pcList})
}

// embeddedHTML 是一个函数，它接受一个namespace字符串并返回完整的HTML字符串
func embeddedHTML(namespacesStr string) string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Web Form and File List</title>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
    margin: 0;
    padding: 0;
    background-color: #f5f5f7;
    color: #2c3e50;
    display: flex;
    justify-content: center;
  }
  .container {
    display: flex;
    width: 100%;
    max-width: 1200px;
  }
  .form-container {
    width: 20%;
    padding: 20px;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
  }
  .file-list {
    width: 80%;
    padding: 20px;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
  }
  .form-container h2, .file-list h2 {
    text-align: center;
  }
  .form-container form {
    display: flex;
    flex-direction: column;
  }
  .form-container label {
    margin-bottom: 5px;
    font-weight: bold;
  }
  .form-container select, .form-container input {
    padding: 10px;
    margin-bottom: 15px;
    border: 1px solid #dfe1e5;
    border-radius: 4px;
  }
  .form-container button {
    padding: 10px;
    border: none;
    border-radius: 4px;
    background-color: #007aff;
    color: white;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  .form-container button:hover {
    background-color: #0050a0;
  }
  .file-list ul {
    list-style: none;
    padding: 0;
  }
  .file-list li {
    margin-bottom: 10px;
  }
  .file-list a {
    text-decoration: none;
    color: #007aff;
    font-weight: bold;
  }
  .file-list a:hover {
    text-decoration: underline;
  }
</style>
</head>
<body>
  <div class="container">
    <div class="form-container">
      <h2>Submit Form</h2>
      <form id="form">
        <label for="namespace">Namespace:</label>
        <select id="namespace" name="namespace">
          <option value="">Select Namespace</option>
        </select><br><br>
        <label for="pod">POD:</label>
        <select id="pod" name="pod">
          <option value="">Select POD</option>
        </select><br><br>
        <label for="container">Container:</label>
        <select id="container" name="container">
          <option value="">Select Container</option>
        </select><br><br>
        <button type="submit">Submit</button>
      </form>
    </div>
    <div class="file-list">
      <h2>File List</h2>
      <ul id="fileList"></ul>
    </div>
  </div>

  <script>
document.addEventListener('DOMContentLoaded', function() {
  // 获取namespace列表并填充到select元素中
  const namespaces = ` + "`" + namespacesStr + "`" + `;
  const namespaceSelect = document.getElementById('namespace');
  namespaces.split(',').forEach(function(namespace) {
    const option = document.createElement('option');
    option.value = namespace.trim();
    option.textContent = namespace.trim();
    namespaceSelect.appendChild(option);
  });
	// 监听namespace的变化，更新pod列表
	namespaceSelect.addEventListener('change', function() {
	  const namespace = this.value;
	  fetch('/iteminfo?namespace=' + encodeURIComponent(namespace))
		.then(response => response.json())
		.then(data => {
		  const podSelect = document.getElementById('pod');
		  podSelect.innerHTML = '<option value="">Select POD</option>'; // 清空pod列表
		  data.viewPod.forEach(pc => {
			const option = document.createElement('option');
			option.value = pc.podname;
			option.textContent = pc.podname;
			podSelect.appendChild(option);
		  });
		})
		.catch((error) => {
		  console.error('Error:', error);
		});
	});

	// 监听pod的变化，更新container列表
	document.getElementById('pod').addEventListener('change', function() {
	  const pod = this.value;
	  const namespace = document.getElementById('namespace').value;
	  fetch('/iteminfo?namespace=' + encodeURIComponent(namespace))
		.then(response => response.json())
		.then(data => {
		  const containerSelect = document.getElementById('container');
		  containerSelect.innerHTML = '<option value="">Select Container</option>'; // 清空容器列表
		  data.viewPod.forEach(pc => {
			if (pc.podname === pod) {
			  pc.containers.forEach(container => {
				const option = document.createElement('option');
				option.value = container;
				option.textContent = container;
				containerSelect.appendChild(option);
			  });
			}
		  });
		})
		.catch((error) => {
		  console.error('Error:', error);
		});
	});

  document.getElementById('form').addEventListener('submit', function(e) {
    e.preventDefault();
    const namespace = document.getElementById('namespace').value;
    const pod = document.getElementById('pod').value;
    const container = document.getElementById('container').value;
    const data = { namespace, pod, container };

    fetch('/', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    .then(response => response.json())
    .then(data => {
      console.log('Success:', data);
    })
    .catch((error) => {
      console.error('Error:', error);
    });
  });

  function fetchFiles() {
    fetch('/files')
      .then(response => response.json())
      .then(data => {
        const fileList = document.getElementById('fileList');
        fileList.innerHTML = '';
        data.forEach(file => {
          const listItem = document.createElement('li');
          const link = document.createElement('a');
          link.href = file;
          link.textContent = file;
          listItem.appendChild(link);
          fileList.appendChild(listItem);
        });
      })
      .catch((error) => {
        console.error('Error:', error);
      });
  }

  fetchFiles();
});
</script>
</body>
</html>`
}
