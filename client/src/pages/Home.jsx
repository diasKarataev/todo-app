import React, { useState, useEffect } from 'react';
import axios from 'axios';
import '../App.css';
import { RxPencil2 } from "react-icons/rx";
import { RxCross1 } from "react-icons/rx";
import { RxStarFilled } from "react-icons/rx";
import { RxStar } from "react-icons/rx";
import { MdKeyboardArrowLeft } from "react-icons/md";
import { MdKeyboardArrowRight } from "react-icons/md";
import Navbar from "../components/Navbar";

const Home = () => {
  const [tasks, setTasks] = useState([]);
  const [newTask, setNewTask] = useState({ name: '', details: '' });
  const [editingTask, setEditingTask] = useState(null);
  const [pagination, setPagination] = useState({ page: 1, pageSize: 5 });
  const [filters, setFilters] = useState({
    name: '',
    details: '',
    star: false,
    sortField: '',
    sortOrder: 'asc',
  });
  const [isAccountActivated, setIsAccountActivated] = useState(false);

  const getToken = () => {
    return localStorage.getItem('token');
  };

  useEffect(() => {
    if (pagination.page !== 1 || tasks.length === 0) {
      fetchTasks();
    }
    fetchUserInfo();
  }, [pagination.page]);


  const fetchUserInfo = async () => {
    try {
      const response = await axios.get('http://localhost:8000/api/user-info', {
        headers: {
          Authorization: `Bearer ${getToken()}`,
        },
      });
      setIsAccountActivated(response.data.isActivated);
    } catch (error) {
      console.error('Ошибка получения информации о пользователе:', error);
    }
  };

  const fetchTasks = async () => {
    try {
      const response = await axios.get('http://localhost:8000/api/tasks', {
        params: {
          page: pagination.page,
          pageSize: pagination.pageSize,
          ...filters,
          sortOrder: filters.sortOrder,
        },
        headers: {
          Authorization: `Bearer ${getToken()}`, // Добавление токена к заголовкам запроса
        },
      });

      setTasks(response.data);
      setPagination((prev) => ({ ...prev, total: response.headers['x-total-count'] }));
    } catch (error) {
      console.error('Ошибка получения задач:', error);
    }
  };

  const handleFilterChange = (event) => {
    const { name, value, type } = event.target;

    const newValue = type === 'checkbox' ? !filters[name] : value;

    setFilters((prevFilters) => ({
      ...prevFilters,
      [name]: newValue,
    }));
  };

  const handleSortChange = (sortField) => {
    setFilters((prevFilters) => ({
      ...prevFilters,
      sortField,
    }));
  };

  const handleSortOrderChange = (sortOrder) => {
    setFilters((prevFilters) => ({
      ...prevFilters,
      sortOrder,
    }));
  };

  const handleInputChange = (event) => {
    const { name, value } = event.target;
    setNewTask({ ...newTask, [name]: value });
  };
  const handleEdit = (task) => {
    setEditingTask(task);
  };

  const handleUpdate = async () => {
    try {
      const response = await axios.put(`http://localhost:8000/api/tasks/${editingTask.ID}`, editingTask, {
        headers: {
          Authorization: `Bearer ${getToken()}`,
          'Content-Type': 'application/json',
        },
      });
      const updatedTasks = tasks.map((task) => (task.ID === editingTask.ID ? response.data : task));
      setTasks(updatedTasks);
      setEditingTask(null);
    } catch (error) {
      console.error('Ошибка обновления задачи:', error);
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();

    try {
      const response = await axios.post('http://localhost:8000/api/tasks', newTask, {
        headers: {
          Authorization: `Bearer ${getToken()}`, // Добавление токена к заголовкам запроса
          'Content-Type': 'application/json', // Указание типа содержимого для запроса
        },
      });
      setTasks([...tasks, response.data]);
      setNewTask({ name: '', details: '' });
    } catch (error) {
      console.error('Ошибка создания задачи:', error);
    }
  };

  const handleEditedInputChange = (event) => {
    const { name, value } = event.target;
    setEditingTask({ ...editingTask, [name]: value });
  };

  const handleCancelEdit = () => {
    setEditingTask(null);
  };

  const handleToggleStar = async (taskId) => {
    try {
      const response = await axios.put(`http://localhost:8000/api/tasks/${taskId}/toggle-star`, null, {
        headers: {
          Authorization: `Bearer ${getToken()}`, // Добавление токена к заголовкам запроса
          'Content-Type': 'application/json', // Указание типа содержимого для запроса
        },
      });
      const { haveStar, lastUpdated } = response.data;

      setTasks((prevTasks) =>
          prevTasks.map((task) =>
              task.ID === taskId ? { ...task, star: haveStar, lastUpdated } : task
          )
      );
    } catch (error) {
      console.error('Ошибка изменения статуса звезды:', error);
    }
  };


  const handleFilterApply = () => {
    fetchTasks();
  };

  const handleDelete = async (taskId) => {
    try {
      await axios.delete(`http://localhost:8000/api/tasks/${taskId}`, {
        headers: {
          Authorization: `Bearer ${getToken()}`, // Добавление токена к заголовкам запроса
        },
      });
      const updatedTasks = tasks.filter((task) => task.ID !== taskId);
      setTasks(updatedTasks);
    } catch (error) {
      console.error('Ошибка удаления задачи:', error);
    }
  };

  const handlePageChange = (newPage) => {
    setPagination(prev => ({ ...prev, page: newPage }));
  };

  return (
      <>
        <Navbar></Navbar>
      <div className='container'>
        {isAccountActivated ? (
            <>
            <h1>Добавить задачу</h1>
            <form onSubmit={handleSubmit}>
              <input
                  type="text"
                  name="name"
                  placeholder="Название задачи"
                  value={newTask.name}
                  onChange={handleInputChange}
              />
              <input
                  type="text"
                  name="details"
                  placeholder="Детали"
                  value={newTask.details}
                  onChange={handleInputChange}
              />
              <button type="submit">Добавить задачу</button>
            </form>
            </>
        ) : (
            <h1 style={{color: "red"}}>Чтобы добавить задачу активируйте аккаунт</h1>
        )}


        <h1>Список задач</h1>

        {/* Filter controls */}
        <div className='filtering'>
          <label>
            Name:
            <input type='text' name='name' value={filters.name} onChange={handleFilterChange}/>
          </label>
          <label>
            Details:
            <input type='text' name='details' value={filters.details} onChange={handleFilterChange}/>
          </label>
          <label>
            Star:
            <input type='checkbox' name='star' checked={filters.star} onChange={handleFilterChange}/>
          </label>
          <p></p>

          <p></p>
          {/* Sorting controls */}
          <label>
            Sort by:
            <div className='select-container'>
            <select name='sortField' value={filters.sortField} onChange={(e) => handleSortChange(e.target.value)}>
              <option value=''>None</option>
              <option value='name'>Name</option>
              <option value='details'>Details</option>
              <option value='lastUpdated'>Updated time</option>
            </select>
            </div>
            <select className='sort-order' name='sortOrder' value={filters.sortOrder} onChange={(e) => handleSortOrderChange(e.target.value)}>
              <option value='asc'>По возрастанию</option>
              <option value='desc'>По убыванию</option>

            </select>
            <div class='search-button-container'>
              <button onClick={handleFilterApply}>Поиск</button>
            </div>
            
          </label>

          <ul>
            {tasks.map((task) => (
                <li key={task.ID}>
                  {editingTask && editingTask.ID === task.ID ? (
                      <div>
                        <input type="text" name="name" value={editingTask.name} onChange={handleEditedInputChange}/>
                        <input type="text" name="details" value={editingTask.details}
                               onChange={handleEditedInputChange}/>
                        <button onClick={handleUpdate}>Сохранить</button>
                        <button className='cancel-btn' onClick={handleCancelEdit}>Отмена</button>
                      </div>
                  ) : (
                      <div className='edit-section'>
                        <div className="task-details">
                          <h2>{task.name}</h2>
                          <p>{task.details}</p>
                        </div>
                        <div className="icon-buttons">
                          <RxPencil2 className='edit-btn' onClick={() => handleEdit(task)} />
                          <RxCross1 className='delete-btn' onClick={() => handleDelete(task.ID)} />
                          {task.star ? (
                            <RxStarFilled className='star-btn starred' onClick={() => handleToggleStar(task.ID)} />
                          ) : (
                            <RxStar className='star-btn' onClick={() => handleToggleStar(task.ID)} />
                          )}
                        </div>
                      </div>
                  )}
                </li>
            ))}
          </ul>

          {/* Pagination controls */}
          <div className='pagination'>
          <MdKeyboardArrowLeft
            className='arrow-l'
            onClick={() => handlePageChange(pagination.page - 1)}
            style={{ visibility: pagination.page === 1 ? 'hidden' : 'visible' }}
          />
            <span>{pagination.page}</span>
            <MdKeyboardArrowRight
              className='arrow-r'
              onClick={() => handlePageChange(pagination.page + 1)}
              disabled={pagination.page === Math.ceil(pagination.total / pagination.pageSize)}
            />
          </div>
        </div>
      </div>
      </>
  );
};

export default Home;
