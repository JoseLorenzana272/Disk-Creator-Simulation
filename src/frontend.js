// Interpreter.js
import React, {useState, useRef} from 'react';
import './styles.css';

function Interpreter() {
    const [code, setCode] = useState('');
    const [output, setOutput] = useState('');
    const fileInputRef = useRef(null);

    const handleOpenFile = () => {
        fileInputRef.current.click();
    };

    const handleFileChange = (event) => {
        const file = event.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (e) => {
                setCode(e.target.result); // Se pone el contenido del archivo de entrada
            };
            reader.readAsText(file);

            // Restablecer el valor del input de archivo
            event.target.value = null;
        }
    };

    const handleNewFile = () => {
        setCode('');
        setOutput('');
    };

    const handleHelp = () => {
        alert('This is a simple code interpreter. Write your code in the input section and click "Run" to execute it.');
    };

    const handleRun = async () => {
        try {
            const response = await fetch('http://localhost:8080/run-code', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ code }),
            });
    
            if (!response.ok) {
                // Leer el cuerpo como texto para incluir en el mensaje de error
                const errorText = await response.text();
                throw new Error(`Error al ejecutar el código: ${response.status} ${response.statusText}\n${errorText}`);
            }
    
            // Leer el cuerpo como JSON
            const result = await response.json();
            const formattedOutput = result.output.map(obj => obj.message || JSON.stringify(obj)).join('\n');
            setOutput(formattedOutput);
        } catch (error) {
            setOutput(`Error: ${error.message}`);
        }
    };
    

    const handleGitHub = () => {
        window.open('https://github.com/JoseLorenzana272', '_blank');
    };


    return (
        <div className="interpreter">
            <div className="sidebar">
                <h1><span className="tag">&lt;</span><span className="name">EXT2 sim.</span><span
                    className="tag">/&gt;</span></h1>
                <button id="openButton" className="cssbuttons-io" onClick={handleOpenFile}>
          <span>
            <svg className="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg"
                 width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
              <path fill-rule="evenodd"
                    d="M9 2.221V7H4.221a2 2 0 0 1 .365-.5L8.5 2.586A2 2 0 0 1 9 2.22ZM11 2v5a2 2 0 0 1-2 2H4v11a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2h-7Zm-.293 9.293a1 1 0 0 1 0 1.414L9.414 14l1.293 1.293a1 1 0 0 1-1.414 1.414l-2-2a1 1 0 0 1 0-1.414l2-2a1 1 0 0 1 1.414 0Zm2.586 1.414a1 1 0 0 1 1.414-1.414l2 2a1 1 0 0 1 0 1.414l-2 2a1 1 0 0 1-1.414-1.414L14.586 14l-1.293-1.293Z"
                    clip-rule="evenodd"/>
            </svg>
            Open File
          </span>
                </button>

                <input
                    type="file"
                    ref={fileInputRef}
                    style={{display: 'none'}}
                    onChange={handleFileChange}
                />

                <button id="newButton" className="cssbuttons-io" onClick={handleNewFile}>
          <span>
            {/* SVG for New File */}
              <svg className="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true"
                   xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
              <path fillRule="evenodd"
                    d="M9 7V2.221a2 2 0 0 0-.5.365L4.586 6.5a2 2 0 0 0-.365.5H9Zm2 0V2h7a2 2 0 0 1 2 2v6.41A7.5 7.5 0 1 0 10.5 22H6a2 2 0 0 1-2-2V9h5a2 2 0 0 0 2-2Z"
                    clipRule="evenodd"/>
              <path fillRule="evenodd"
                    d="M9 16a6 6 0 1 1 12 0 6 6 0 0 1-12 0Zm6-3a1 1 0 0 1 1 1v1h1a1 1 0 1 1 0 2h-1v1a1 1 0 1 1-2 0v-1h-1a1 1 0 1 1 0-2h1v-1a1 1 0 0 1 1-1Z"
                    clipRule="evenodd"/>
            </svg>
            New File
          </span>
                </button>

                <button id="helpButton" className="cssbuttons-io" onClick={handleHelp}>
          <span>
            {/* SVG for Help */}
              <svg className="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true"
                   xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
              <path fillRule="evenodd"
                    d="M2 12C2 6.477 6.477 2 12 2s10 4.477 10 10-4.477 10-10 10S2 17.523 2 12Zm9.008-3.018a1.502 1.502 0 0 1 2.522 1.159v.024a1.44 1.44 0 0 1-1.493 1.418 1 1 0 0 0-1.037.999V14a1 1 0 1 0 2 0v-.539a3.44 3.44 0 0 0 2.529-3.256 3.502 3.502 0 0 0-7-.255 1 1 0 0 0 2 .076c.014-.398.187-.774.48-1.044Zm.982 7.026a1 1 0 1 0 0 2H12a1 1 0 1 0 0-2h-.01Z"
                    clipRule="evenodd"/>
            </svg>
            Help
          </span>
                </button>

                <button id="GithubButton" className="buttonGit" onClick={handleGitHub}>
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path
                            d="M12 0.296997C5.37 0.296997 0 5.67 0 12.297C0 17.6 3.438 22.097 8.205 23.682C8.805 23.795 9.025 23.424 9.025 23.105C9.025 22.82 9.015 22.065 9.01 21.065C5.672 21.789 4.968 19.455 4.968 19.455C4.422 18.07 3.633 17.7 3.633 17.7C2.546 16.956 3.717 16.971 3.717 16.971C4.922 17.055 5.555 18.207 5.555 18.207C6.625 20.042 8.364 19.512 9.05 19.205C9.158 18.429 9.467 17.9 9.81 17.6C7.145 17.3 4.344 16.268 4.344 11.67C4.344 10.36 4.809 9.29 5.579 8.45C5.444 8.147 5.039 6.927 5.684 5.274C5.684 5.274 6.689 4.952 8.984 6.504C9.944 6.237 10.964 6.105 11.984 6.099C13.004 6.105 14.024 6.237 14.984 6.504C17.264 4.952 18.269 5.274 18.269 5.274C18.914 6.927 18.509 8.147 18.389 8.45C19.154 9.29 19.619 10.36 19.619 11.67C19.619 16.28 16.814 17.295 14.144 17.59C14.564 17.95 14.954 18.686 14.954 19.81C14.954 21.416 14.939 22.706 14.939 23.096C14.939 23.411 15.149 23.786 15.764 23.666C20.565 22.092 24 17.592 24 12.297C24 5.67 18.627 0.296997 12 0.296997Z"
                            fill="white"/>
                    </svg>
                    <p className="text">GitHub</p>
                </button>
            </div>

            <div className="container">
                <div className="input-section">
                    <h2><span className="tag">&lt;</span><span className="name">Input.</span><span className="tag">/&gt;</span></h2>
                    <textarea id="inputCode"
                              placeholder="Write your code here..."
                              value={code}
                              onChange={(e) => setCode(e.target.value)}
                    ></textarea>

                    <button id="runButton" className="cssbuttons-io" onClick={handleRun}>
            <span>
              <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path d="M0 0h24v24H0z" fill="none"></path>
                  <path
                      d="M24 12l-5.657 5.657-1.414-1.414L21.172 12l-4.243-4.243 1.414-1.414L24 12zM2.828 12l4.243 4.243-1.414 1.414L0 12l5.657-5.657L7.07 7.757 2.828 12zm6.96 9H7.66l6.552-18h2.128L9.788 21z"
                      fill="currentColor"
                  ></path>
                </svg>
              Run
            </span>
                    </button>
                </div>
                <div className="output-section">
                    <h2><span className="tag">&lt;</span><span className="name">Output.</span><span
                        className="tag">/&gt;</span></h2>
                    <textarea
                        id="output"
                        readOnly
                        value={output || '# Here you\'ll see the results of execution.'}
                    />

                </div>
            </div>
        </div>
    );
}

export default Interpreter;
