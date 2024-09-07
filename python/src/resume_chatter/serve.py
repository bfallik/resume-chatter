import logging

from langchain.chains import create_retrieval_chain
from langchain.chains.combine_documents import create_stuff_documents_chain
from langchain_community.document_loaders import PyPDFLoader
from langchain_community.vectorstores import FAISS
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.runnables.base import Runnable
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_text_splitters import RecursiveCharacterTextSplitter

logger = logging.getLogger(__name__)


from concurrent import futures

import grpc
from protocgenpy.chat.v1 import chat_pb2, chat_pb2_grpc


class ChatService(chat_pb2_grpc.ChatServiceServicer):
    def Ask(self, request, context):
        model = "gpt-4o-mini"
        logger.warning(f"using model: {model}")

        llm = ChatOpenAI(model=model)

        loader = PyPDFLoader(request.document_path)
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

        return chat_pb2.AskResponse(
            response=str(retrieval_chain.invoke({"input": request.question}))
        )


def serve() -> None:
    port = "8081"
    # GRPC enables SO_REUSEPORT by default.
    #   https://groups.google.com/g/grpc-io/c/RB69llv2tC4
    # This may be recommended for production workloads but causes problems
    # during development when I inevitably launch a second server without
    # killing the first. Disable this behavior for my sanity.
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10), options=(("grpc.so_reuseport", 0),)
    )
    chat_pb2_grpc.add_ChatServiceServicer_to_server(ChatService(), server)
    server.add_insecure_port("[::]:" + port)
    server.start()
    print("Server started, listening on " + port)
    server.wait_for_termination()
