BEGIN;

UPDATE candidates SET resume_text = 'Sarah Chen
Senior Software Engineer

Experience:
Google - Senior Software Engineer (2021-Present)
- Led development of distributed systems for Google Cloud Platform
- Designed and implemented microservices architecture using Go and gRPC
- Managed team of 5 engineers, mentored 2 junior developers
- Reduced API latency by 40% through caching and query optimization

Stripe - Software Engineer (2018-2021)
- Built payment processing pipelines handling millions of daily transactions
- Developed fraud detection algorithms using Python and machine learning
- Implemented real-time monitoring dashboards with React and TypeScript

Education:
Stanford University - MS Computer Science (2018)
UC Berkeley - BS Computer Science (2016)

Skills: Go, Python, Java, TypeScript, React, Kubernetes, Docker, AWS, GCP, PostgreSQL, Redis, gRPC, microservices, distributed systems, machine learning, TensorFlow'
WHERE email = 'sarah.chen@gmail.com';

UPDATE candidates SET resume_text = 'James Rodriguez
Staff Frontend Engineer

Experience:
Meta - Staff Frontend Engineer (2020-Present)
- Architected React component library used across 15+ internal products
- Led migration from legacy PHP templates to React/GraphQL
- Built accessibility framework achieving WCAG 2.1 AA compliance
- Performance optimized News Feed rendering, improving Core Web Vitals by 35%

Airbnb - Senior Frontend Engineer (2017-2020)
- Developed Airbnb Luxe booking experience using React Native
- Built design system with 200+ reusable components
- Implemented server-side rendering reducing time-to-interactive by 50%

Education:
MIT - BS Computer Science (2017)

Skills: React, TypeScript, JavaScript, GraphQL, Next.js, React Native, CSS, HTML, Webpack, Node.js, accessibility, performance optimization, design systems, A/B testing'
WHERE email = 'james.rodriguez@outlook.com';

UPDATE candidates SET resume_text = 'Emily Watson
Backend Engineer

Experience:
Stripe - Backend Engineer (2021-Present)
- Built billing infrastructure supporting subscription management for enterprise clients
- Designed event-driven architecture using Kafka and PostgreSQL
- Implemented webhook delivery system with guaranteed at-least-once delivery
- Wrote comprehensive API documentation and developer guides

Twilio - Software Engineer (2019-2021)
- Developed REST APIs for SMS and voice communication services
- Built rate limiting and throttling systems using Redis
- Contributed to open-source SDK libraries in Python and Ruby

Education:
Carnegie Mellon University - BS Computer Science (2019)

Skills: Python, Ruby, Go, PostgreSQL, Kafka, Redis, REST APIs, event-driven architecture, Docker, AWS, CI/CD, technical writing'
WHERE email = 'emily.watson@protonmail.com';

UPDATE candidates SET resume_text = 'Raj Patel
Principal Engineer

Experience:
Amazon - Principal Engineer (2019-Present)
- Designed and built AWS Lambda cold start optimization reducing startup time by 60%
- Led architecture review board for AWS Serverless division
- Filed 8 patents related to cloud computing and serverless architecture
- Mentored 20+ engineers across multiple teams

Microsoft - Senior Software Engineer (2015-2019)
- Core contributor to Azure Functions runtime
- Built auto-scaling infrastructure handling 100K+ concurrent requests
- Developed infrastructure-as-code tools using Terraform and ARM templates

Education:
Georgia Tech - MS Computer Science (2015)
IIT Bombay - BTech Computer Science (2013)

Skills: Java, C#, Python, Go, AWS, Azure, serverless, Lambda, Terraform, Kubernetes, Docker, system design, distributed computing, patents, cloud architecture'
WHERE email = 'raj.patel@gmail.com';

UPDATE candidates SET resume_text = 'Maria Gonzalez
Data Engineer

Experience:
Netflix - Senior Data Engineer (2020-Present)
- Built real-time recommendation pipeline processing 500M+ events daily
- Designed data lake architecture on AWS S3 with Apache Spark
- Implemented ML feature store used by 10+ data science teams
- Reduced data pipeline costs by 30% through partition optimization

Uber - Data Engineer (2018-2020)
- Developed ETL pipelines for rider and driver analytics using Apache Airflow
- Built real-time surge pricing data pipeline with Apache Flink
- Managed Hadoop clusters processing 50TB+ daily

Education:
University of Washington - MS Data Science (2018)
UCLA - BS Statistics (2016)

Skills: Python, Scala, SQL, Apache Spark, Kafka, Airflow, Flink, Hadoop, AWS, S3, Redshift, Snowflake, dbt, data modeling, ETL, machine learning, statistics'
WHERE email = 'maria.gonzalez@yahoo.com';

UPDATE candidates SET resume_text = 'Michael Park
DevOps / SRE Lead

Experience:
Microsoft - Senior SRE Lead (2019-Present)
- Led SRE team of 8 managing 99.99% uptime for Azure Active Directory
- Built observability platform using Prometheus, Grafana, and OpenTelemetry
- Automated incident response reducing MTTR from 45min to 12min
- Implemented chaos engineering practices across 3 critical services

Dropbox - Site Reliability Engineer (2016-2019)
- Managed Kubernetes clusters serving 600M+ users
- Built CI/CD pipelines deploying 200+ times per day
- Designed disaster recovery procedures and ran regular game days

Education:
University of Michigan - BS Computer Engineering (2016)

Skills: Kubernetes, Docker, Terraform, AWS, Azure, GCP, Prometheus, Grafana, Python, Go, Bash, Linux, CI/CD, Jenkins, GitHub Actions, incident management, chaos engineering, SRE'
WHERE email = 'michael.park@gmail.com';

UPDATE candidates SET resume_text = 'Priya Sharma
iOS Engineer

Experience:
Apple - Senior iOS Engineer (2020-Present)
- Core contributor to Apple Maps navigation engine
- Built SwiftUI components for iOS 17 Weather app redesign
- Optimized battery consumption reducing power usage by 25%
- Implemented Core ML models for on-device text recognition

Lyft - iOS Engineer (2017-2020)
- Developed rider app features serving 30M+ monthly active users
- Built offline-first architecture using Core Data and CloudKit
- Implemented driver ETA prediction with on-device machine learning

Education:
University of Toronto - MS Computer Science (2017)
Delhi University - BTech Information Technology (2015)

Skills: Swift, SwiftUI, Objective-C, UIKit, Core Data, Core ML, Xcode, iOS, watchOS, Combine, RxSwift, MVVM, unit testing, XCTest, accessibility, performance profiling'
WHERE email = 'priya.sharma@gmail.com';

UPDATE candidates SET resume_text = 'Daniel Kim
Machine Learning Engineer

Experience:
Databricks - Senior ML Engineer (2021-Present)
- Built AutoML pipeline reducing model training time by 70%
- Developed LLM fine-tuning infrastructure on multi-GPU clusters
- Implemented MLflow model registry serving 500+ production models
- Published 3 papers on efficient transformer architectures

OpenAI - ML Engineer (2019-2021)
- Contributed to GPT-3 training infrastructure
- Built evaluation frameworks for language model benchmarking
- Developed RLHF data collection and annotation tools

Education:
Stanford University - PhD Computer Science, specialization in NLP (2019)
KAIST - BS Computer Science (2014)

Skills: Python, PyTorch, TensorFlow, JAX, CUDA, transformers, NLP, LLMs, deep learning, distributed training, MLOps, MLflow, Spark, Databricks, GPU optimization, research'
WHERE email = 'daniel.kim@icloud.com';

UPDATE candidates SET resume_text = 'Lauren Mitchell
Product Designer / UX Engineer

Experience:
Notion - Senior Product Designer (2021-Present)
- Designed database views feature used by 4M+ teams
- Led design system evolution with 150+ Figma components
- Conducted 50+ user research sessions informing product roadmap
- Built interactive prototypes using React and Framer Motion

Figma - UX Engineer (2019-2021)
- Developed design-to-code tooling for FigJam collaboration features
- Built plugin architecture supporting 1000+ community plugins
- Created accessibility audit tools for design file review

Education:
Rhode Island School of Design - BFA Graphic Design (2019)
Minor in Computer Science

Skills: Figma, Sketch, React, TypeScript, CSS, HTML, Framer Motion, user research, prototyping, design systems, accessibility, information architecture, wireframing, usability testing'
WHERE email = 'lauren.mitchell@gmail.com';

UPDATE candidates SET resume_text = 'Ahmed Hassan
Security Engineer

Experience:
Airbnb - Senior Security Engineer (2020-Present)
- Led application security program for payment processing systems
- Built automated vulnerability scanning pipeline catching 200+ issues pre-production
- Designed and implemented zero-trust network architecture
- Conducted security training for 500+ engineers

CrowdStrike - Security Researcher (2017-2020)
- Discovered 12 CVEs in major enterprise software
- Built threat intelligence platform processing 10M+ indicators daily
- Developed YARA rules and Sigma detections for APT groups
- Presented research at DEF CON and Black Hat conferences

Education:
NYU Tandon - MS Cybersecurity (2017)
Cairo University - BS Computer Engineering (2015)

Skills: Python, Go, Rust, penetration testing, threat modeling, OWASP, Burp Suite, AWS security, IAM, SOC2, SIEM, incident response, cryptography, reverse engineering, vulnerability research'
WHERE email = 'ahmed.hassan@outlook.com';

UPDATE candidates SET resume_text = 'Jessica Liu
Full Stack Engineer

Experience:
Figma - Full Stack Engineer (2021-Present)
- Built real-time collaboration engine supporting 1000+ concurrent editors
- Developed REST and WebSocket APIs using Rust and TypeScript
- Optimized canvas rendering achieving 60fps with 10K+ objects
- Implemented conflict resolution algorithms for multiplayer editing

Vercel - Software Engineer (2019-2021)
- Core contributor to Next.js framework (ISR, middleware features)
- Built deployment pipeline serving 100K+ sites
- Developed edge functions runtime for serverless computing

Education:
UC Berkeley - BS EECS (2019)

Skills: TypeScript, Rust, React, Next.js, Node.js, WebSockets, PostgreSQL, Redis, GraphQL, WebGL, Canvas API, real-time systems, distributed systems, performance optimization'
WHERE email = 'jessica.liu@gmail.com';

UPDATE candidates SET resume_text = 'Tom Anderson
Data Platform Engineer

Experience:
Snowflake - Senior Data Platform Engineer (2020-Present)
- Built query optimizer improvements reducing average query time by 25%
- Designed multi-cluster warehouse auto-scaling algorithms
- Implemented time-travel and data sharing features
- Led cross-functional team of 6 engineers and 2 PMs

Palantir - Forward Deployed Engineer (2017-2020)
- Built data integration pipelines for Fortune 500 clients
- Developed geospatial analytics platform for defense applications
- Implemented ontology modeling system for complex data relationships

Education:
Columbia University - MS Computer Science (2017)
University of Colorado - BS Computer Science (2015)

Skills: Java, Python, Scala, SQL, Snowflake, Spark, Kafka, Flink, data warehousing, query optimization, distributed databases, AWS, Terraform, data modeling, geospatial analytics'
WHERE email = 'tom.anderson@gmail.com';

UPDATE candidates SET resume_text = 'Nina Petrov
Backend / Systems Engineer

Experience:
Palantir - Senior Systems Engineer (2020-Present)
- Designed secure multi-tenant platform serving government and enterprise clients
- Built high-throughput data ingestion pipelines handling 1TB+ daily
- Implemented end-to-end encryption for data at rest and in transit
- Achieved FedRAMP High authorization for cloud deployment

Cloudflare - Systems Engineer (2017-2020)
- Developed edge computing runtime handling 25M+ requests per second
- Built DDoS mitigation system using eBPF and XDP
- Optimized TLS handshake performance reducing latency by 30%

Education:
ETH Zurich - MS Computer Science (2017)
Moscow State University - BS Applied Mathematics (2015)

Skills: Rust, C++, Go, Python, Linux, networking, TCP/IP, eBPF, security, cryptography, distributed systems, Kubernetes, performance engineering, systems programming, FedRAMP'
WHERE email = 'nina.petrov@protonmail.com';

UPDATE candidates SET resume_text = 'Kevin O''Brien
Blockchain / Web3 Engineer

Experience:
Coinbase - Senior Software Engineer (2021-Present)
- Built smart contract auditing tools for ERC-20 and ERC-721 tokens
- Developed custody infrastructure managing $50B+ in digital assets
- Implemented Layer 2 scaling solutions using Optimistic Rollups
- Built real-time transaction monitoring and compliance systems

ConsenSys - Software Engineer (2019-2021)
- Core contributor to MetaMask browser extension
- Developed Solidity smart contracts for DeFi protocols
- Built Ethereum node infrastructure handling 1M+ daily transactions

Education:
Georgia Tech - MS Computer Science (2019)
Boston University - BS Computer Science (2017)

Skills: Solidity, TypeScript, Rust, Go, Ethereum, smart contracts, DeFi, Web3, React, Node.js, PostgreSQL, Redis, cryptography, blockchain architecture, security auditing'
WHERE email = 'kevin.obrien@gmail.com';

UPDATE candidates SET resume_text = 'Fatima Al-Rashid
Platform / Infrastructure Engineer

Experience:
Datadog - Senior Platform Engineer (2020-Present)
- Built metric ingestion pipeline processing 10 trillion data points daily
- Designed auto-scaling infrastructure reducing cloud costs by 40%
- Implemented distributed tracing system with sub-millisecond overhead
- Led migration from monolith to 50+ microservices architecture

HashiCorp - Software Engineer (2018-2020)
- Core contributor to Terraform AWS provider
- Built Vault secrets management integrations for Kubernetes
- Developed Consul service mesh features for multi-datacenter deployments

Education:
University of Texas at Austin - MS Computer Science (2018)
American University of Beirut - BE Computer Engineering (2016)

Skills: Go, Python, Rust, Kubernetes, Docker, Terraform, Vault, Consul, AWS, GCP, Prometheus, Datadog, distributed systems, observability, infrastructure-as-code, microservices, SRE'
WHERE email = 'fatima.alrashid@gmail.com';

COMMIT;
