import {useState} from "react";

function App() {

  const [hasUploaded, setUploaded] = useState(false);

  return (
    <>
      <h1>Borel Budget</h1>
      <p>
        Upload your budget spreadsheet and we'll analyse it
      </p>
      {hasUploaded ?
        <p>Thanks for uploading your file</p>
        : <form method="POST" action='http://127.0.0.1:3001/csvupload' className='upload-form' encType="multipart/form-data">
          <input type='file' name="file" required accept='text/csv'/>
          <button className="submit" type='submit'>Upload</button>
        </form>
      }
    </>
  )
}

export default App
