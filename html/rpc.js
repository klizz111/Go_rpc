/**
 * 调用RPC服务器的RpcRunCommand方法
 * @param {string} command - 要执行的命令
 * @returns {Promise<string>} - 命令执行的结果
 */
async function callRpcRunCommand(command) {
    // 构造RPC请求体，符合JSON-RPC规范
    const rpcRequest = {
        jsonrpc: "2.0",
        id: Date.now(),
        method: "Call.RpcRunCommand",
        params: [command, ""] // 添加第二个参数作为结果占位符
    };

    console.log('发送RPC请求:', JSON.stringify(rpcRequest));
    
    try {
        // 发送POST请求到RPC服务器
        const response = await fetch('http://localhost:1234/rpc', {
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