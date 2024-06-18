console.log("打开首页")

//	改变AI开启状态
function changeAiEnableStatus(wxId) {
    // console.log("修改AI开启状态: ", wxId)

    axios({
        method: 'put',
        url: '/api/ai/status',
        data: {
            wxId: wxId
        }
    }).then(function (response) {
        console.log(`返回结果: ${JSON.stringify(response)}`);
        alert(`${response.data}`)
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
        alert("修改失败")
    }).finally(function () {
        window.location.reload();
    })
}

// 修改水群排行榜状态
function changeGroupRankEnableStatus(wxId) {
    // console.log("修改水群排行榜开启状态: ", wxId)
    axios({
        method: 'put',
        url: '/api/grouprank/status',
        data: {
            wxId: wxId
        }
    }).then(function (response) {
        console.log(`返回结果: ${JSON.stringify(response)}`);
        alert(`${response.data}`)
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
        alert("修改失败")
    }).finally(function () {
        window.location.reload();
    })
}

// 修改水群排行榜状态
function changeSummaryEnableStatus(wxId) {
    // console.log("修改聊天记录总结开启状态: ", wxId)
    axios({
        method: 'put',
        url: '/api/summary/status',
        data: {
            wxId: wxId
        }
    }).then(function (response) {
        console.log(`返回结果: ${JSON.stringify(response)}`);
        alert(`${response.data}`)
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
        alert("修改失败")
    }).finally(function () {
        window.location.reload();
    })
}

// 修改欢迎语开启状态
function changeWelcomeEnableStatus(wxId) {
    axios({
        method: 'put',
        url: '/api/welcome/status',
        data: {
            wxId: wxId
        }
    }).then(function (response) {
        console.log(`返回结果: ${JSON.stringify(response)}`);
        alert(`${response.data}`)
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
        alert("修改失败")
    }).finally(function () {
        window.location.reload();
    })
}

// 修改用户新闻开启状态
function changeUserNewsStatus(wxId) {
  axios({
    method: 'put',
    url: '/api/news/status',
    data: {
      wxId: wxId
    }
  }).then(function (response) {
    console.log(`返回结果: ${JSON.stringify(response)}`);
    alert(`${response.data}`)
  }).catch(function (error) {
    console.log(`错误信息: ${error}`);
    alert("修改失败")
  }).finally(function () {
    window.location.reload();
  })
}

// 修改指令权限启用状态
function changeCommandEnableStatus(wxId) {
    axios({
        method: 'put',
        url: '/api/command/status',
        data: {
            wxId: wxId
        }
    }).then(function (response) {
        console.log(`返回结果: ${JSON.stringify(response)}`);
        alert(`${response.data}`)
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
        alert("修改失败")
    }).finally(function () {
        window.location.reload();
    })
}

// 修改群成员是否参与排行榜状态
function changeUserGroupRankSkipStatus(groupId, userId) {
    console.log("修改水群排行榜开启状态: ", groupId, userId)
    axios({
        method: 'put',
        url: '/api/grouprank/skip',
        data: {
            wxId: groupId,
            userId: userId
        }
    }).then(function (response) {
        console.log(`返回结果: ${JSON.stringify(response)}`);
        alert(`${response.data}`)
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
        alert("修改失败")
    }).finally(function () {
        window.location.reload();
    })
}

// 获取群成员列表
function getGroupUsers(groupId, groupName) {
    // 获取表格的tbody部分，以便稍后向其中添加行
    var tbody = document.getElementById("groupUsers");
    tbody.innerHTML = ""

    // 打开模态框
    const modal = document.getElementById("groupUserModal");
    modal.showModal()

    // 设置群名称
    const groupNameTag = document.getElementById("groupUserModalName");
    groupNameTag.innerHTML = '<span class="loading loading-dots loading-lg"></span>'

    // 显示加载框
    // const loading = document.getElementById("groupUserDataLoading");
    // loading.style.display = "block"

    axios.get('/api/group/users', {
        params: {
            groupId: groupId
        }
    }).then(function (response) {
        // console.log(`返回结果: ${JSON.stringify(response)}`);
        // 渲染群成员列表
        const groupUsers = response.data
        // 循环渲染数据
        groupUsers.forEach((groupUser, i) => {
            console.log(groupUser)
            const { wxid, nickname, isMember, isAdmin, joinTime, lastActive, leaveTime, skipChatRank } = groupUser;

            let row = tbody.insertRow(i);
            // Insert data into cells
            row.insertCell(0).innerHTML = wxid;
            row.insertCell(1).innerHTML = nickname;
            row.insertCell(2).innerHTML = `<span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${isMember ? 'bg-green-50 text-green-700 ring-green-600/20' : 'bg-red-50 text-red-700 ring-red-600/20'}">${isMember ? '是' : '否'}</span>`;
            row.insertCell(3).innerHTML = `<span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${isAdmin ? 'bg-green-50 text-green-700 ring-green-600/20' : 'bg-red-50 text-red-700 ring-red-600/20'}">${isAdmin ? '是' : '否'}</span>`;
            row.insertCell(4).innerHTML = joinTime;
            row.insertCell(5).innerHTML = lastActive;
            row.insertCell(6).innerHTML = leaveTime;
            // row.insertCell(7).innerHTML = `<input type="checkbox" class="toggle toggle-error" ${skipChatRank ? 'checked' : ''} onclick="changeUserGroupRankSkipStatus('${groupId}', '${wxid}')" />`;
            row.insertCell(7).innerHTML = `<span class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${skipChatRank ? 'bg-green-50 text-green-700 ring-green-600/20' : 'bg-red-50 text-red-700 ring-red-600/20'}">${skipChatRank ? '是' : '否'}</span>`;
        });
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
    }).finally(function () {
        // 隐藏加载框
        // loading.style.display = "none"
        groupNameTag.innerHTML = groupName
    })
}

// AI模型变动
function aiModelChange(event, wxid) {
    // 取出变动后的值
    const modelStr = event.target.value;
    console.log("AI模型变动: ", wxid, modelStr)
    axios({
        method: 'post',
        url: '/api/ai/model',
        data: {
            wxid: wxid,
            model: modelStr
        }
    }).then(function (response) {
        console.log(`返回结果: ${JSON.stringify(response)}`);
        alert(`${response.data}`)
    }).catch(function (error) {
        console.log(`错误信息: ${error}`);
        alert("修改失败")
    }).finally(function () {
        window.location.reload();
    })
}

// AI角色变动
function aiAssistantChange(event, wxid) {
  // 取出变动后的值
  const assistantStr = event.target.value;
  console.log("AI角色变动: ", wxid, assistantStr)
  axios({
    method: 'post',
    url: '/api/ai/assistant',
    data: {
      wxid: wxid,
      model: assistantStr
    }
  }).then(function (response) {
    console.log(`返回结果: ${JSON.stringify(response)}`);
    alert(`${response.data}`)
  }).catch(function (error) {
    console.log(`错误信息: ${error}`);
    alert("修改失败")
  }).finally(function () {
    window.location.reload();
  })
}

// 修改清理群成员值
function changeClearMember(wxid, oldVal, newVal) {
  oldVal = Number(oldVal)
  newVal = Number(newVal)

  if (oldVal === newVal) {
    return
  }
  if (newVal < 0) {
    alert('不活跃天数的值不能小于0')
  }
  // 请求接口
  axios({
    method: 'put',
    url: '/api/clearmembers',
    data: {
      wxid: wxid,
      days: Number(newVal)
    }
  }).then(function (response) {
    console.log(`返回结果: ${JSON.stringify(response)}`);
    alert(`${response.data}`)
  }).catch(function (error) {
    console.log(`错误信息: ${error}`);
    alert("修改失败")
  }).finally(function () {
    window.location.reload();
  })
}
