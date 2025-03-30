/**
 * 调用RPC服务器的RpcRunCommand方法
 * @param {string} command - 要执行的命令
 * @returns {Promise<string>} - 命令执行的结果
 */
async function callRpcRunCommand(command) {
    // 构造RPC请求体，符合JSON-RPC 2.0规范
    const authcode = document.getElementById('authcode').value;
    localStorage.setItem('authcode', authcode);
    const dest = document.getElementById('dest').value;
    localStorage.setItem('dest', dest);
    const rpcRequest = {
        jsonrpc: "2.0",
        id: Date.now(),
        method: "Call.RpcRunCommand",
        params: [command], // 只传递一个参数，服务器会自动处理
        authcode: authcode
    };

    console.log('发送RPC请求:', JSON.stringify(rpcRequest));
    url = 'http://' + dest + '/rpc';
    try {
        // 发送POST请求到RPC服务器
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(rpcRequest)
        });

        // 打印原始响应文本以便调试
        const responseText = await response.text();
        console.log('原始响应:', responseText);
        
        // 检查响应状态
        if (!response.ok) {
            throw new Error(`HTTP error: ${response.status} - ${response.statusText}`);
        }

        try {
            // 尝试解析JSON
            const result = JSON.parse(responseText);
            
            // 检查RPC错误
            if (result.error) {
                throw new Error(`RPC error: ${JSON.stringify(result.error)}`);
            }
            
            return result.result;
        } catch (parseError) {
            console.error('解析响应失败:', parseError);
            return responseText; // 返回原始文本作为备选
        }
    } catch (error) {
        console.error('RPC调用失败:', error);
        throw error;
    }
}

/**
 * 使用示例
 */
function executeCommand() {
    const commandInput = document.getElementById('commandInput');
    localStorage.setItem('command', commandInput.value);
    const resultOutput = document.getElementById('resultOutput');
    
    if (!commandInput || !commandInput.value.trim()) {
        alert('请输入要执行的命令');
        return;
    }
    
    resultOutput.textContent = '执行中...';
    
    callRpcRunCommand(commandInput.value.trim())
        .then(result => {
            resultOutput.textContent = result;
        })
        .catch(error => {
            resultOutput.textContent = `错误: ${error.message}`;
        });
}

async function executeCode() {
    const resultOutput = document.getElementById('resultOutput');
    const codeInput = document.getElementById('shortcode').value;
    const dest = document.getElementById('dest').value;
    localStorage.setItem('dest', dest);
    url = 'http://' + dest + '/code';

    if (codeInput == "") {
        alert('请输入要执行的代码');
        return;
    }
    localStorage.setItem('code', codeInput.value);
    const codeRequest = {
        shortcode: codeInput
    }
    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(codeRequest)
        });
        const responseText = await response.text();
        console.log('原始响应:', responseText);
        if (!response.ok) {
            throw new Error(`HTTP error: ${response.status} - ${response.statusText}`);
        }
        try {
            const result = JSON.parse(responseText);
            if (result.error) {
                throw new Error(`RPC error: ${JSON.stringify(result.error)}`);
            }
            resultOutput.textContent = result.result;
            return result.result;
        } catch (parseError) {
            console.error('解析响应失败:', parseError);
            resultOutput.textContent = responseText;
            return
        } 
    }
    catch (error) {
        console.error('RPC调用失败:', error);
        resultOutput.textContent = `错误: ${error.message}`;
    }

}