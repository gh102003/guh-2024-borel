import { useEffect, useState } from "react";
import Markdown from 'react-markdown'

function App() {

  const urlSearchParams = new URLSearchParams(window.location.search);
  const params = Object.fromEntries(urlSearchParams.entries());
  const hasUploaded = params.uploaded === "true";

  const [analysis, setAnalysis] = useState<string | null>(null);

  useEffect(() => {
    if (hasUploaded) {
      fetch("/api/getinsight/123")
        .then(res => res.text())
        .then(res => setAnalysis(res.split("Response from OpenAI:")[1].trim()));
    } else {
      setAnalysis(null);
    }
  }, [hasUploaded]);

  return (
    <>
      <h1>Borel Budget</h1>
      <p>
        Upload your budget spreadsheet and we'll analyse it
      </p>
      {hasUploaded ?
        <p>Thanks for uploading your file</p>
        : <form method="POST" action='http://127.0.0.1:3001/csvupload' className='upload-form' encType="multipart/form-data">
          <input type='file' name="file" required accept='text/csv' />
          <button className="submit" type='submit'>Upload</button>
        </form>
      }
      {hasUploaded && !analysis &&
        <p>
          Loading...
        </p>
      }
      {analysis &&
        <div className="analysis">
          <p>Response from OpenAI:</p>
          <p className="analysis-text">
            <Markdown>{analysis}</Markdown>
          </p>
        </div>
      }
    </>
  )
}

export default App
