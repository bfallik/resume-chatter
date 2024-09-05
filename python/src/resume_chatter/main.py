from langchain.chains import create_retrieval_chain
from langchain.chains.combine_documents import create_stuff_documents_chain
from langchain_community.document_loaders import PyPDFLoader
from langchain_community.vectorstores import FAISS
from langchain_core.prompts import ChatPromptTemplate
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_text_splitters import RecursiveCharacterTextSplitter

import logging

logger = logging.getLogger(__name__)


if __name__=="__main__":
    model = "gpt-4o-mini"
    logger.warning(f"using model: {model}")

    llm = ChatOpenAI(model=model)

    loader = PyPDFLoader("/home/bfallik/Documents/JobSearches/bfallik-resume/bfallik-resume.pdf")
    docs = loader.load()

    text_splitter = RecursiveCharacterTextSplitter()
    documents = text_splitter.split_documents(docs)
    embeddings = OpenAIEmbeddings()
    vector = FAISS.from_documents(documents, embeddings)

    prompt = ChatPromptTemplate.from_template("""Answer the following question based only on the provided context:

    <context>
    {context}
    </context>

    Question: {input}""")

    document_chain = create_stuff_documents_chain(llm, prompt)
    retriever = vector.as_retriever()
    retrieval_chain = create_retrieval_chain(retriever, document_chain)

    res = retrieval_chain.invoke({"input": "What was Brian's second most recent job and when did he work there?"})
    print(res)
